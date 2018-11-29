package plugin

import (
	"errors"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/pb"
)

var (
	plugins []Plugin
)

const (
	TYPE_ANALYSIS int8 = iota
	TYPE_RERANK
)

type Plugin interface {
	Type() int8
	Name() string
	Priority() int8
	Init() error
	Analyze(args *xsearchpb.QueryAnalysisArgs) error
	Rerank(args *xsearchpb.RerankArgs) error
	Cleanup() error
}

func Init(sodir string) error {
	err := filepath.Walk(
		sodir,
		func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if !strings.HasSuffix(info.Name(), ".so") {
				return nil
			}
			plg, err := plugin.Open(path)
			if err != nil {
				log.Error(err)
				return err
			}
			symbol, err := plg.Lookup("Object")
			if err != nil {
				log.Error(err)
				return err
			}
			object, ok := symbol.(Plugin)
			if !ok {
				err := errors.New("object convert fail.")
				log.Error(err)
				return err
			}
			if err := object.Init(); err != nil {
				log.Error(err)
				return err
			}
			plugins = append(plugins, object)
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
