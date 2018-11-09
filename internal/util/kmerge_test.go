package util

import (
	"bytes"
	"testing"
)

const (
	K   = 5
	MIN = ""
	MAX = "y0"
)

type foo struct {
	v [][]interface{}
	i []int
}

func TestKMerge(t *testing.T) {
	f := &foo{
		v: make([][]interface{}, K),
		i: make([]int, K),
	}
	f.v[0] = []interface{}{interface{}("a"), interface{}("f"), interface{}("k"), interface{}("p"), interface{}("u")}
	f.v[1] = []interface{}{interface{}("-1"), interface{}("g"), interface{}("l"), interface{}("q"), interface{}("v")}
	f.v[2] = []interface{}{interface{}("-2"), interface{}("h"), interface{}("m"), interface{}("r"), interface{}("w")}
	f.v[3] = []interface{}{interface{}("d"), interface{}("i"), interface{}("n"), interface{}("s"), interface{}("x")}
	f.v[4] = []interface{}{interface{}("e"), interface{}("j"), interface{}("o"), interface{}("t"), interface{}("y")}
	input := func(i int) interface{} {
		if i >= K {
			return interface{}(MAX)
		}
		if f.i[i] < K {
			k := f.v[i][f.i[i]]
			f.i[i]++
			return k
		} else {
			return interface{}(MAX)
		}
	}
	output := func(k interface{}) {
		println(k.(string))
	}
	compare := func(a, b interface{}) int {
		return bytes.Compare([]byte(a.(string)), []byte(b.(string)))
	}
	KMerge(K, interface{}(MIN), interface{}(MAX), input, output, compare)
	println()
}
