package preprocessor

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"unicode"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
	"github.com/tddhit/xsearch/plugin"
)

func init() {
	if err := plugin.Register(&preprocessor{}); err != nil {
		log.Fatal(err)
	}
}

type preprocessor struct {
	t2s       map[rune]rune //繁转简
	stopwords map[rune]struct{}
	initial   bool
}

func (p *preprocessor) Init() error {
	p.initial = true
	p.t2s = make(map[rune]rune)
	p.stopwords = make(map[rune]struct{})
	for i := range trad {
		p.t2s[trad[i]] = simp[i]
	}
	file, err := os.Open("./stopwords.dict")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key := scanner.Text()
		p.stopwords[[]rune(key)[0]] = struct{}{}
	}
	return nil
}

func (p *preprocessor) Type() int8 {
	return plugin.TYPE_ANALYSIS
}

func (p *preprocessor) Name() string {
	return "preprocessor"
}

func (p *preprocessor) Priority() int8 {
	return 1
}

func (p *preprocessor) Analyze(args *xsearchpb.QueryAnalysisArgs) error {
	if !p.initial {
		return nil
	}
	query := args.Queries[0].Raw
	query = strings.Replace(query, "\\n", "\n", -1)
	query = strings.Replace(query, "\n", "", -1)
	query = strings.Trim(query, " ")
	query = strQ2B(query)
	query = p.toSimp(query)
	words := []rune(query)
	buf := bytes.Buffer{}
	for _, word := range words {
		if _, ok := p.stopwords[word]; ok {
			continue
		} else if isUnchar(word) {
			buf.WriteRune(word)
		} else if word == ' ' {
			buf.WriteRune(word)
		}
	}
	args.Queries[0].Raw = buf.String()
	return nil
}

func (p *preprocessor) Rerank(args *xsearchpb.RerankArgs) error {
	return nil
}

func (p *preprocessor) Cleanup() error {
	return nil
}

func strQ2B(ustring string) string {
	var rstring string
	ustring_ := []rune(ustring)
	for _, uchar := range ustring_ {
		if uchar == 12288 {
			uchar = 32
		} else if uchar >= 65281 && uchar <= 65374 {
			uchar -= 65248
		}
		uchar = unicode.ToLower(uchar)
		rstring += string(uchar)
	}
	return rstring
}

func isUnchar(uchar rune) bool {
	if unicode.Is(unicode.Scripts["Han"], uchar) {
		return true
	}
	if unicode.IsDigit(uchar) {
		return true
	}
	if unicode.IsLower(uchar) {
		return true
	}
	return false
}
func (p *preprocessor) toSimp(s string) string {
	r := []rune(s)
	for i := range r {
		if _, ok := p.t2s[r[i]]; ok {
			r[i] = p.t2s[r[i]]
		}
	}
	return string(r)
}
