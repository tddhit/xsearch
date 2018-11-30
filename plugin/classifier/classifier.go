package plugin

import (
	"bufio"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/tensorflow/tensorflow/tensorflow/go"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
)

const (
	unknownWord = 'U'
	paddingWord = 'P'
)

func init() {
	var (
		c   *Classifier
		err error
	)
	if c, err = newClassifier(); err != nil {
		log.Fatal(err)
	}
	if err = plugin.Register(c); err != nil {
		log.Fatal(err)
	}
}

type Classifier struct {
	session       *tensorflow.Session
	graph         *tensorflow.Graph
	tags          []string
	embedding     map[rune][]float32
	embeddingSize int
	maxLength     int
}

func newClassifier() (*Classifier, error) {
	data, err := ioutil.ReadFile("./classifier.model")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	graph := tensorflow.NewGraph()
	if err := graph.Import(data, ""); err != nil {
		log.Error(err)
		return nil, err
	}
	session, err := tensorflow.NewSession(graph, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	tags, err := loadTags("./tags.txt")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	embedding, err := loadEmbedding("./embedding.txt")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	c := &Classifier{
		graph:         graph,
		session:       session,
		tags:          tags,
		embedding:     embedding,
		embeddingSize: 32,
		maxLength:     197,
	}
	return c, nil
}

func (c *Classifier) Type() int8 {
	return plugin.TYPE_ANALYSIS
}

func (c *Classifier) Name() string {
	return "classifier"
}

func (c *Classifier) Priority() int8 {
	return 2
}

func (c *Classifier) Analyze(args *xsearchpb.QueryAnalysisArgs) error {
	words := []rune(args.Queries[0].Raw)
	embedding := c.getEmbedding(words)
	matrix := [][][]float32{embedding}
	input, err := tensorflow.NewTensor(matrix)
	if err != nil {
		log.Error(err)
		return err
	}
	dropout, _ := tensorflow.NewTensor(float32(1.0))
	feeds := map[tensorflow.Output]*tensorflow.Tensor{
		c.graph.Operation("input_x").Output(0):           input,
		c.graph.Operation("dropout_keep_prob").Output(0): dropout,
	}
	output, err := c.session.Run(feeds,
		[]tensorflow.Output{
			c.graph.Operation("output/scores_softmax").Output(0),
		},
		nil,
	)
	if err != nil {
		log.Error(err)
		return err
	}
	scores := output[0].Value().([][]float32)[0]
	for _, i := range topKIndex(scores, 5) {
		args.Queries[0].Tags = append(args.Queries[0].Tags, c.tags[i])
	}
	return nil
}

func topKIndex(scores []float32, k int) []int {
	type kv struct {
		k float32
		v int
	}
	kvs := make([]kv, len(scores))
	res := make([]int, k)
	for i, score := range scores {
		kvs[i] = kv{score, i}
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].k > kvs[j].k
	})
	for i := range res {
		res[i] = kvs[i].v
	}
	return res
}

func (c *Classifier) Rerank(args *xsearchpb.RerankArgs) error {
	return nil
}

func (c *Classifier) Cleanup() error {
	if err := c.session.Close(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *Classifier) getEmbedding(words []rune) [][]float32 {
	embedding := make([][]float32, c.maxLength)
	for i := range embedding {
		embedding[i] = make([]float32, c.embeddingSize)
	}
	for i, word := range words {
		if _, ok := c.embedding[word]; ok {
			embedding[i] = c.embedding[word]
		} else {
			embedding[i] = c.embedding[unknownWord]
		}
	}
	for i := len(words); i < c.maxLength; i++ {
		embedding[i] = c.embedding[paddingWord]
	}
	return embedding
}

func loadEmbedding(path string) (map[rune][]float32, error) {
	embedding := make(map[rune][]float32)
	file, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	tempIndex := 0 //跳过第一行（词的个数和词向量维度）
	for scanner.Scan() {
		if tempIndex > 0 {
			values := strings.Split(scanner.Text(), " ")
			word := []rune(values[0])[0]
			vec := values[1:]
			embedding[word] = make([]float32, len(vec))
			for i, _ := range vec {
				v, err := strconv.ParseFloat(vec[i], 32)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				embedding[word][i] = float32(v)
			}
		}
		tempIndex += 1
	}
	return embedding, nil
}

func loadTags(path string) (tags []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tags = append(tags, scanner.Text())
	}
	return
}
