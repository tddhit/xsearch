package indexer

import "time"

type indexerOptions struct {
	persistNum      uint64
	persistInterval time.Duration
	mergeInterval   time.Duration
	dir             string
	id              string
	shardNum        int
}

type IndexerOption func(*indexerOptions)

func WithPersistNum(n int) IndexerOption {
	return func(o *indexerOptions) {
		o.persistNum = uint64(n)
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

func WithShardNum(num int) IndexerOption {
	return func(o *indexerOptions) {
		o.shardNum = num
	}
}
