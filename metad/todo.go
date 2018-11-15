package metad

import (
	"fmt"
	"sync"

	"github.com/tddhit/tools/log"

	"github.com/tddhit/xsearch/metad/pb"
)

type action int

const (
	ACTION_CREATE_SHARD action = iota
	ACTION_MIGRATE_SHARD
)

func (a action) String() string {
	switch a {
	case ACTION_CREATE_SHARD:
		return "create shard"
	case ACTION_MIGRATE_SHARD:
		return "migrate shard"
	default:
		return "unknown"
	}
}

type todo struct {
	action action
	from   *node
	to     *node
	shard  *shard
	table  *shardTable
	stages []func() error
	wg     *sync.WaitGroup
}

func newTodo(a action, from, to *node, s *shard) {
	d := &todo{
		action: a,
		from:   from,
		to:     to,
		shard:  s,
	}
	s.todo = d
	switch d.action {
	case ACTION_CREATE_SHARD:
		d.buildCreateStages()
	case ACTION_MIGRATE_SHARD:
		d.buildMigrateStages()
	}
}

func (d *todo) getInfo() string {
	if d == nil {
		return ""
	}
	switch d.action {
	case ACTION_CREATE_SHARD:
		return d.action.String() + ":" + d.from.addr
	case ACTION_MIGRATE_SHARD:
		return fmt.Sprintf("%s: %s -> %s", d.action.String(), d.from.addr, d.to.addr)
	default:
		return "unknown"
	}
}

func (d *todo) do(wg ...*sync.WaitGroup) (int, error) {
	if d.wg == nil && len(wg) == 1 {
		d.wg = wg[0]
	}
	if len(d.stages) == 0 {
		return 0, nil
	}
	stageFunc := d.stages[0]
	if err := stageFunc(); err != nil {
		log.Error(err)
		return len(d.stages), err
	}
	d.stages = d.stages[1:]
	if len(d.stages) == 0 {
		d.wg.Done()
		d.wg = nil
		return 0, nil
	}
	return len(d.stages), nil
}

func (d *todo) buildCreateStages() {
	d.stages = append(d.stages, func() error {
		d.from.writeC <- &metadpb.RegisterNodeRsp{
			Type:    metadpb.RegisterNodeRsp_CreateShard,
			ShardID: d.shard.id,
		}
		return nil
	})

	d.stages = append(d.stages, func() error {
		d.from.status = NODE_CLUSTER_ONLINE
		d.shard.node = d.from
		d.shard.next = nil
		return nil
	})
}

func (d *todo) buildMigrateStages() {
	d.stages = append(d.stages, func() error {
		d.to.writeC <- &metadpb.RegisterNodeRsp{
			Type:    metadpb.RegisterNodeRsp_CreateShard,
			ShardID: d.shard.id,
		}
		return nil
	})

	d.stages = append(d.stages, func() error {
		d.to.status = NODE_CLUSTER_ONLINE
		d.shard.node = d.to
		d.shard.next = nil
		d.from.writeC <- &metadpb.RegisterNodeRsp{
			Type:    metadpb.RegisterNodeRsp_RemoveShard,
			ShardID: d.shard.id,
		}
		return nil
	})
}
