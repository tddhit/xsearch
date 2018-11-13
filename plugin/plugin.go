package plugin

import (
	"errors"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"github.com/tddhit/tools/log"
)

var (
	iplugins []*iplugin
)

type PluginType int

const (
	PLUGIN_ANALYSIS PluginType = iota
	PLUGIN_RERANK
)

type iplugin struct {
	name     string
	typ      PluginType
	priority int8
	init     func() error
	process  func(args interface{}) error
	cleanup  func() error
}

func initSO(sodir string) error {
	err := filepath.Walk(
		sodir,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".so") {
				plg, err := plugin.Open(path)
				if err != nil {
					return err
				}
				nameSymbol, err := plg.Lookup("Name")
				if err != nil {
					return err
				}
				_, ok := nameSymbol.(func() string)
				if !ok {
					return errors.New("Name() convert fail.")
				}
				typeSymbol, err := plg.Lookup("Type")
				if err != nil {
					return err
				}
				prioritySymbol, err := plg.Lookup("Priority")
				if err != nil {
					return err
				}
				initSymbol, err := plg.Lookup("Init")
				if err != nil {
					return err
				}
				processSymbol, err := plg.Lookup("Process")
				if err != nil {
					return err
				}
				cleanupSymbol, err := plg.Lookup("Cleanup")
				if err != nil {
					return err
				}
				p := &iplugin{
					name:     nameSymbol.(func() string)(),
					typ:      typeSymbol.(func() PluginType)(),
					priority: prioritySymbol.(func() int8)(),
					init:     initSymbol.(func() error),
					process:  processSymbol.(func(args interface{}) error),
					cleanup:  cleanupSymbol.(func() error),
				}
				iplugins = append(iplugins, p)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	sort.SliceStable(iplugins, func(i, j int) bool {
		return iplugins[i].priority < iplugins[j].priority
	})
	return nil
}

func analyze(args *QueryArgs) error {
	for _, p := range iplugins {
		if p.typ != PLUGIN_ANALYSIS {
			continue
		}
		if err := p.process(args); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func rerank(args *DocsArgs) error {
	for _, p := range iplugins {
		if p.typ != PLUGIN_RERANK {
			continue
		}
		if err := p.process(args); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func cleanup() error {
	for _, p := range iplugins {
		if err := p.cleanup(); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
