package indexer

import "time"

type indexerOptions struct {
	CommitNumDocs      uint64
	CommitTimeInterval time.Duration
	IndexDir           string
	IndexID            int
	Sharding           int
}

type IndexerOption func(*indexerOptions)

func WithCommitNumDocs(n int) IndexerOption {
	return func(o *indexerOptions) {
		o.CommitNumDocs = uint64(n)
	}
}

func WithCommitTimeInverval(t time.Duration) IndexerOption {
	return func(o *indexerOptions) {
		o.CommitTimeInterval = t
	}
}

func WithIndexDir(path string) IndexerOption {
	return func(o *indexerOptions) {
		o.IndexDir = path
	}
}

func WithIndexID(id int) IndexerOption {
	return func(o *indexerOptions) {
		o.IndexID = id
	}
}

func WithSharding(num int) IndexerOption {
	return func(o *indexerOptions) {
		o.Sharding = num
	}
}
