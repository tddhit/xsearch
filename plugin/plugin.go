package plugin

import (
	"errors"
	"fmt"
	"sort"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
)

var (
	plugins []Plugin
	m       = make(map[string]Plugin)
)

const (
	TYPE_ANALYSIS int8 = iota
	TYPE_RERANK
)

type Plugin interface {
	Init() error
	Type() int8
	Name() string
	Priority() int8
	Analyze(args *xsearchpb.QueryAnalysisArgs) error
	Rerank(args *xsearchpb.RerankArgs) error
	Cleanup() error
}

func Register(p Plugin) error {
	if _, ok := m[p.Name()]; ok {
		err := fmt.Errorf("plugin(%s) already exist.", p.Name())
		log.Error(err)
		return err
	}
	m[p.Name()] = p
	plugins = append(plugins, p)
	sort.SliceStable(plugins, func(i, j int) bool {
		return plugins[i].Priority() < plugins[j].Priority()
	})
	return nil
}

func Get(name string) (Plugin, bool) {
	p, ok := m[name]
	return p, ok
}

func Analyze(args *xsearchpb.QueryAnalysisArgs) error {
	if args == nil || len(args.Queries) == 0 {
		err := errors.New("there is no query.")
		log.Error(err)
		return err
	}
	for _, p := range plugins {
		if p.Type() != TYPE_ANALYSIS {
			continue
		}
		if err := p.Analyze(args); err != nil {
			return err
		}
	}
	return nil
}

func Rerank(args *xsearchpb.RerankArgs) error {
	for _, p := range plugins {
		if p.Type() != TYPE_RERANK {
			continue
		}
		if err := p.Rerank(args); err != nil {
			return err
		}
	}
	return nil
}

func Cleanup() error {
	sort.SliceStable(plugins, func(i, j int) bool {
		return plugins[i].Priority() > plugins[j].Priority()
	})
	for _, p := range plugins {
		if err := p.Cleanup(); err != nil {
			return err
		}
	}
	return nil
}
