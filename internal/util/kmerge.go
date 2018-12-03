package util

import "github.com/tddhit/tools/log"

func KMerge(
	k int,
	input func(int) (interface{}, error),
	output func(interface{}) error,
	cmp func(a, b interface{}) int) (err error) {

	if k == 0 {
		return
	}
	t := &loserTree{
		k:      k,
		cmp:    cmp,
		leaves: make([]interface{}, k),
		losers: make([]int, k),
	}
	var (
		over bool
		end  int
	)
	for i := 0; i < t.k; i++ {
		if over, err = kInput(t, &end, input, i); err != nil {
			log.Error(err)
			return
		} else if over {
			return
		}
	}
	t.build()
	for {
		if err = output(t.leaves[t.losers[0]]); err != nil {
			log.Error(err)
			return
		}
		if over, err = kInput(t, &end, input, t.losers[0]); err != nil {
			log.Error(err)
			return
		} else if over {
			return
		}
		t.adjust(t.losers[0])
	}
	return
}

type loserTree struct {
	k      int
	cmp    func(a, b interface{}) int
	leaves []interface{}
	losers []int
}

func (t *loserTree) build() {
	for i := range t.losers {
		t.losers[i] = -1
	}
	for i := t.k - 1; i >= 0; i-- {
		t.adjust(i)
	}
}

func (t *loserTree) adjust(i int) {
	for parent := (i + t.k) / 2; parent > 0; parent /= 2 {
		if t.losers[parent] == -1 {
			t.losers[parent], i = i, -1
			continue
		}
		if t.leaves[t.losers[parent]] == nil {
			continue
		}
		if i == -1 {
			continue
		}
		if t.leaves[i] == nil {
			t.losers[parent], i = i, t.losers[parent]
			continue
		}
		if t.cmp(t.leaves[i], t.leaves[t.losers[parent]]) > 0 {
			t.losers[parent], i = i, t.losers[parent]
			continue
		}
	}
	t.losers[0] = i
}

func kInput(
	t *loserTree,
	end *int,
	f func(int) (interface{}, error),
	i int) (over bool, err error) {

	var value interface{}
	if value, err = f(i); err != nil {
		log.Error(err)
		return
	} else {
		if value == nil {
			*end++
			if *end == t.k {
				over = true
				return
			}
		}
	}
	t.leaves[i] = value
	return
}
