package util

func KMerge(
	k int, min, max interface{},
	input func(i int) interface{},
	output func(k interface{}),
	compare func(a, b interface{}) int) {

	t := &loserTree{
		k:       k,
		min:     min,
		max:     max,
		input:   input,
		output:  output,
		compare: compare,
		b:       make([]interface{}, k+1),
		ls:      make([]int, k),
	}
	t.merge()
}

type loserTree struct {
	k       int
	min     interface{}
	max     interface{}
	input   func(i int) interface{}
	output  func(k interface{})
	compare func(a, b interface{}) int
	b       []interface{}
	ls      []int
}

func (t *loserTree) build() {
	t.b[t.k] = t.min
	for i := range t.ls {
		t.ls[i] = t.k
	}
	for i := t.k - 1; i >= 0; i-- {
		t.adjust(i)
	}
}

func (t *loserTree) adjust(s int) {
	for i := (s + t.k) / 2; i > 0; i /= 2 {
		if t.compare(t.b[s], t.b[t.ls[i]]) > 0 {
			t.ls[i] ^= s
			s ^= t.ls[i]
			t.ls[i] ^= s
		}
	}
	t.ls[0] = s
}

func (t *loserTree) merge() {
	for i := 0; i < t.k; i++ {
		t.b[i] = t.input(i)
	}
	t.build()
	for t.compare(t.b[t.ls[0]], t.max) != 0 {
		q := t.ls[0]
		t.output(t.b[q])
		t.b[q] = t.input(q)
		t.adjust(q)
	}
}
