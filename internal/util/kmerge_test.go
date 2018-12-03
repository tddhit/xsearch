package util

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/tddhit/tools/log"
)

func assert(b bool, err string) {
	if !b {
		log.Fatal(err)
	}
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func getNilData() ([][]int, []int) {
	return nil, nil
}

func getEmptyData() ([][]int, []int) {
	return [][]int{}, nil
}

func getEmpty2Data() ([][]int, []int) {
	return make([][]int, 5), nil
}

func getFixedData() ([][]int, []int) {
	return [][]int{{}, {-1, 5, 11}, {}, {6, 10}}, []int{-1, 5, 6, 10, 11}
}

func getRandomData() (data [][]int, orig []int) {
	rand.Seed(time.Now().UnixNano())
	data = make([][]int, 5)
	for i := range data {
		for j := 0; j < rand.Intn(20); j++ {
			data[i] = append(data[i], rand.Intn(100))
		}
		sort.Ints(data[i])
		orig = append(orig, data[i]...)
	}
	sort.Ints(orig)
	return
}

func TestKMergeNilArray(t *testing.T) {
	data, orig := getNilData()
	kMergeArray(data, orig)
}

func TestKMergeEmptyArray(t *testing.T) {
	data, orig := getEmptyData()
	kMergeArray(data, orig)
}

func TestKMergeEmpty2Array(t *testing.T) {
	data, orig := getEmpty2Data()
	kMergeArray(data, orig)
}

func TestKMergeFixedArray(t *testing.T) {
	data, orig := getFixedData()
	kMergeArray(data, orig)
}

func TestKMergeRandomArray(t *testing.T) {
	data, orig := getRandomData()
	kMergeArray(data, orig)
}

func kMergeArray(data [][]int, orig []int) {
	var res []int
	if orig != nil {
		res = make([]int, 0)
	}
	input := func(i int) (interface{}, error) {
		if len(data) == 0 {
			return nil, nil
		}
		if len(data[i]) == 0 {
			return nil, nil
		}
		r := data[i][0]
		data[i] = data[i][1:len(data[i])]
		return r, nil
	}
	output := func(k interface{}) error {
		res = append(res, k.(int))
		return nil
	}
	compare := func(a, b interface{}) int {
		return a.(int) - b.(int)
	}
	KMerge(len(data), input, output, compare)
	assert(equal(orig, res), "kmerge array fail")
}

type ListNode struct {
	Val  int
	Next *ListNode
}

func printLists(lists []*ListNode) {
	for _, head := range lists {
		for head != nil {
			print(head.Val, " ", head.Next, " ")
			head = head.Next
		}
		println()
	}
}

func TestKMergeNilList(t *testing.T) {
	data, orig := getNilData()
	kMergeLinkedList(data, orig)
}

func TestKMergeEmptyList(t *testing.T) {
	data, orig := getEmptyData()
	kMergeLinkedList(data, orig)
}

func TestKMergeEmpty2List(t *testing.T) {
	data, orig := getEmpty2Data()
	kMergeLinkedList(data, orig)
}

func TestKMergeFixedList(t *testing.T) {
	data, orig := getFixedData()
	kMergeLinkedList(data, orig)
}

func TestKMergeRandomList(t *testing.T) {
	data, orig := getRandomData()
	kMergeLinkedList(data, orig)
}

func kMergeLinkedList(data [][]int, orig []int) {
	lists := make([]*ListNode, len(data))
	var (
		res   []int
		dummy = &ListNode{}
		head  = dummy
	)
	for i := range data {
		for j := 0; j < len(data[i]); j++ {
			dummy.Next = &ListNode{Val: data[i][j]}
			dummy = dummy.Next
		}
		lists[i] = head.Next
		dummy = head
		dummy.Next = nil
	}
	input := func(i int) (interface{}, error) {
		if lists[i] != nil {
			head := lists[i]
			lists[i] = lists[i].Next
			return head, nil
		}
		return nil, nil
	}
	output := func(k interface{}) error {
		dummy.Next = k.(*ListNode)
		dummy = dummy.Next
		return nil
	}
	compare := func(a, b interface{}) int {
		return a.(*ListNode).Val - b.(*ListNode).Val
	}
	KMerge(len(data), input, output, compare)
	head = head.Next
	for head != nil {
		res = append(res, head.Val)
		head = head.Next
	}
	assert(equal(orig, res), "kmerge linked list fail")
}
