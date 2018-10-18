package service

import (
	"context"

	"github.com/tddhit/tools/log"

	searchdpb "github.com/tddhit/xsearch/searchd/pb"
)

type action int

const (
	ACTION_CREATE_SHARD action = iota
	ACTION_MIGRATE_SHARD
)

type todo struct {
	id     uint64
	action action
	from   *node
	to     *node
	shard  *shard
	table  *shardTable
	stages []func() error
}

func newTodo(id uint64, a action, from, to *node, s *shard, t *shardTable) *todo {
	d := &todo{
		action: a,
		from:   from,
		to:     to,
		shard:  s,
		table:  t,
	}
	switch d.action {
	case ACTION_CREATE_SHARD:
		d.buildCreateStages()
	case ACTION_MIGRATE_SHARD:
		d.buildMigrateStages()
	}
	return d
}

func (d *todo) do() (count int, err error) {
	if len(d.stages) == 0 {
		return
	}
	stageFunc := d.stages[0]
	if err = stageFunc(); err != nil {
		log.Error(err)
	}
	d.stages = d.stages[1:]
	count = len(d.stages)
	return
}

func (d *todo) buildCreateStages() {
	d.stages = append(d.stages, func() error {
		_, err := d.from.client.CreateShard(
			context.Background(),
			&searchdpb.CreateShardReq{
				ShardID: d.shard.id,
			},
		)
		return err
	})

	d.stages = append(d.stages, func() error {
		d.shard.node = d.from
		d.from.shards = append(d.from.shards, d.shard)
		d.table.addNode(d.from)
		d.table.setShard(d.shard)
		return nil
	})
}

func (d *todo) buildMigrateStages() {
	d.stages = append(d.stages, func() error {
		_, err := d.to.client.CreateShard(
			context.Background(),
			&searchdpb.CreateShardReq{
				ShardID: d.shard.id,
			},
		)
		return err
	})

	d.stages = append(d.stages, func() error {
		d.shard.node = d.to
		d.to.shards = append(d.to.shards, d.shard)
		d.table.addNode(d.to)
		d.table.setShard(d.shard)
		return nil
	})

	d.stages = append(d.stages, func() error {
		_, err := d.from.client.RemoveShard(
			context.Background(),
			&searchdpb.RemoveShardReq{
				ShardID: d.shard.id,
			},
		)
		return err
	})

	d.stages = append(d.stages, func() error {
		return d.from.removeShard(d.shard)
	})
}
