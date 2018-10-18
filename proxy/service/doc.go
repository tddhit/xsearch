package service

import xsearchpb "github.com/tddhit/xsearch/pb"

type DocHeap []*xsearchpb.Document

func (h DocHeap) Len() int           { return len(h) }
func (h DocHeap) Less(i, j int) bool { return h[i].BM25Score > h[j].BM25Score }
func (h DocHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *DocHeap) Push(x interface{}) {
	*h = append(*h, x.(*xsearchpb.Document))
}

func (h *DocHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
