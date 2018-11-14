package plugin

import (
	"errors"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"github.com/tddhit/xsearch/pb"
)

var (
	plugins []Plugin
)

type Type int

const (
	TYPE_ANALYSIS Type = iota
	TYPE_RERANK
)

type Plugin interface {
	Type() Type
	Name() string
	Priority() int8
	Analyze(args *xsearchpb.QueryAnalysisArgs) error
	Rerank(args *xsearchpb.RerankArgs) error
	Cleanup() error
}

func Init(sodir string) error {
	err := filepath.Walk(
		sodir,
		func(path string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(info.Name(), ".so") {
				return nil
			}
			plg, err := plugin.Open(path)
			if err != nil {
				return err
			}
			symbol, err := plg.Lookup("New")
			if err != nil {
				return err
			}
			newSymbol, ok := symbol.(func() (Plugin, error))
			if !ok {
				return errors.New("New symbol convert fail.")
			}
			p, err := newSymbol()
			if err != nil {
				return err
			}
			plugins = append(plugins, p)
			return nil
		},
	)
	if err != nil {
		return err
	}
	sort.SliceStable(plugins, func(i, j int) bool {
		return plugins[i].Priority() < plugins[j].Priority()
	})
	return nil
}

func Analyze(args *xsearchpb.QueryAnalysisArgs) error {
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
