package main

import (
	"io/ioutil"

	"github.com/tensorflow/tensorflow/tensorflow/go"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
)

const (
	unknownWord = 'U'
	paddingWord = 'P'
)

var Object Classifier

type Classifier struct {
	session       *tensorflow.Session
	graph         *tensorflow.Graph
	embedding     map[rune][]float32
	embeddingSize int
	maxLength     int
}

func (c *Classifier) Init() error {
	data, err := ioutil.ReadFile("./classifier.model")
	if err != nil {
		log.Error(err)
		return err
	}
	graph := tensorflow.NewGraph()
	if err := graph.Import(data, ""); err != nil {
		log.Error(err)
		return err
	}
	session, err := tensorflow.NewSession(graph, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	c.graph = graph
	c.session = session
	c.embedding = make(map[rune][]float32)
	c.embeddingSize = 200
	c.maxLength = 300
	return nil
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
	matrix := c.getEmbedding(words)
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
	scores := output[0].Value().([]float32)
	log.Debug(scores)
	return nil
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
