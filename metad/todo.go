package metad

import (
	"github.com/tddhit/tools/log"

	"github.com/tddhit/xsearch/metad/pb"
)

type action int

const (
	ACTION_CREATE_SHARD action = iota
	ACTION_MIGRATE_SHARD
)

type todo struct {
	action action
	from   *node
	to     *node
	shard  *shard
	table  *shardTable
	stages []func() error
}

func newTodo(a action, from, to *node, s *shard) *todo {
	d := &todo{
		action: a,
		from:   from,
		to:     to,
		shard:  s,
	}
	switch d.action {
	case ACTION_CREATE_SHARD:
		d.buildCreateStages()
	case ACTION_MIGRATE_SHARD:
		d.buildMigrateStages()
	}
	return d
}

func (d *todo) do() (int, error) {
	if len(d.stages) == 0 {
		return 0, nil
	}
	stageFunc := d.stages[0]
	if err := stageFunc(); err != nil {
		log.Error(err)
		return len(d.stages), err
	}
	d.stages = d.stages[1:]
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
		d.from.status = NODE_ONLINE
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
		d.from.status = NODE_ONLINE
		d.shard.node = d.to
		d.shard.backup = nil
		d.from.writeC <- &metadpb.RegisterNodeRsp{
			Type:    metadpb.RegisterNodeRsp_RemoveShard,
			ShardID: d.shard.id,
		}
		return nil
	})
}
