package util

import (
	"bytes"
	"testing"
)

const (
	K   = 5
	MIN = "a"
	MAX = "z"
)

type foo struct {
	v [][]key
	i []int
}

func TestKMerge(t *testing.T) {
	f := &foo{
		v: make([][]key, K),
		i: make([]int, K),
	}
	f.v[0] = []key{key("a"), key("f"), key("k"), key("p"), key("u")}
	f.v[1] = []key{key("b"), key("g"), key("l"), key("q"), key("v")}
	f.v[2] = []key{key("c"), key("h"), key("m"), key("r"), key("w")}
	f.v[3] = []key{key("d"), key("i"), key("n"), key("s"), key("x")}
	f.v[4] = []key{key("e"), key("j"), key("o"), key("t"), key("y")}
	input := func(i int) key {
		if f.i[i] < K {
			k := f.v[i][f.i[i]]
			f.i[i]++
			return k
		} else {
			return key(MAX)
		}
	}
	output := func(k key) {
		print(string(k))
	}
	compare := func(a, b key) int {
		return bytes.Compare(a, b)
	}
	KMerge(K, key(MIN), key(MAX), input, output, compare)
	println()
}
