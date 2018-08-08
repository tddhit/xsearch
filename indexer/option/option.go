package option

import "time"

type IndexerOptions struct {
	CommitNumDocs      uint64
	CommitTimeInterval time.Duration
	IndexDir           string
	IndexID            int
	Sharding           int
}

type IndexerOption func(*IndexerOptions)

func WithCommitNumDocs(n int) IndexerOption {
	return func(o *IndexerOptions) {
		o.CommitNumDocs = uint64(n)
	}
}

func WithCommitTimeInverval(t time.Duration) IndexerOption {
	return func(o *IndexerOptions) {
		o.CommitTimeInterval = t
	}
}

func WithIndexDir(path string) IndexerOption {
	return func(o *IndexerOptions) {
		o.IndexDir = path
	}
}

func WithIndexID(id int) IndexerOption {
	return func(o *IndexerOptions) {
		o.IndexID = id
	}
}

func WithSharding(num int) IndexerOption {
	return func(o *IndexerOptions) {
		o.Sharding = num
	}
}
