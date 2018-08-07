package option

import "time"

type IndexerOptions struct {
	CommitNumDocs      int
	CommitTimeInterval time.Duration
	IndexDir           string
	Sharding           int
}

type IndexerOption func(*IndexerOptions)

func WithCommitNumDocs(n int) IndexerOption {
	return func(o *IndexerOptions) {
		o.CommitNumDocs = n
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

func WithSharding(num int) IndexerOption {
	return func(o *IndexerOptions) {
		o.Sharding = num
	}
}
