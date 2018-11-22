package indexer

import "time"

type indexerOptions struct {
	persistInterval time.Duration
	mergeInterval   time.Duration
	mergeNum        uint64
	dir             string
	id              string
}

type IndexerOption func(*indexerOptions)

func WithMergeNum(n int) IndexerOption {
	return func(o *indexerOptions) {
		o.mergeNum = uint64(n)
	}
}

func WithPersistInterval(t time.Duration) IndexerOption {
	return func(o *indexerOptions) {
		o.persistInterval = t
	}
}

func WithMergeInterval(t time.Duration) IndexerOption {
	return func(o *indexerOptions) {
		o.mergeInterval = t
	}
}

func WithDir(dir string) IndexerOption {
	return func(o *indexerOptions) {
		o.dir = dir
	}
}

func WithID(id string) IndexerOption {
	return func(o *indexerOptions) {
		o.id = id
	}
}
