package searchd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/wangbin/jiebago"

	"github.com/tddhit/diskqueue/pb"
	"github.com/tddhit/tools/log"
	"github.com/tddhit/xsearch/indexer"
	"github.com/tddhit/xsearch/pb"
)

type shard struct {
	id        string
	namespace string
	groupID   int
	replicaID int
	indexer   *indexer.Indexer
	segmenter *jiebago.Segmenter
	stopwords map[string]struct{}
	diskq     diskqueuepb.DiskqueueGrpcClient
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func newShard(
	id string,
	dir string,
	addr string,
	segmenter *jiebago.Segmenter,
	stopwords map[string]struct{},
	c diskqueuepb.DiskqueueGrpcClient) (*shard, error) {

	v := strings.Split(id, ".")
	if len(v) != 3 {
		err := fmt.Errorf("invalid shardID:%s", id)
		log.Error(err)
		return nil, err
	}
	namespace := v[0]
	groupID, err := strconv.Atoi(v[1])
	if err != nil {
		err := fmt.Errorf("invalid shardID:%s", id)
		log.Error(err)
		return nil, err
	}
	replicaID, err := strconv.Atoi(v[2])
	if err != nil {
		err := fmt.Errorf("invalid shardID:%s", id)
		log.Error(err)
		return nil, err
	}
	i, err := indexer.New(
		indexer.WithDir(dir),
		indexer.WithID(id),
		indexer.WithMergeInterval(60*time.Second),
		indexer.WithPersistInterval(10*time.Second),
	)
	if err != nil {
		return nil, err
	}
	s := &shard{
		id:        id,
		namespace: namespace,
		groupID:   groupID,
		replicaID: replicaID,
		indexer:   i,
		segmenter: segmenter,
		stopwords: stopwords,
		diskq:     c,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Add(1)
	topic, channel := fmt.Sprintf("%s.%d", s.namespace, s.groupID), addr
	log.Debug(topic, channel)
	go func() {
		s.indexLoop(topic, channel)
		s.wg.Done()
	}()
	return s, nil
}

func (s *shard) indexLoop(topic, channel string) {
	log.Trace(2, "indexLoop start")
	for {
		rsp, err := s.diskq.Pop(s.ctx, &diskqueuepb.PopReq{
			Topic:   topic,
			Channel: channel,
		})
		if err != nil {
			log.Trace(2, "!!!!!!!!!!!!!!!!!!")
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}
		msg := rsp.GetMessage()
		if msg == nil {
			continue
		}
		data := msg.GetData()
		cmd := &xsearchpb.Command{}
		if err := proto.Unmarshal(data, cmd); err != nil {
			log.Error(err)
			continue
		}
		switch cmd.Type {
		case xsearchpb.Command_INDEX:
			if doc, ok := cmd.DocOneof.(*xsearchpb.Command_Doc); ok {
				for term := range s.segmenter.Cut(doc.Doc.Content, true) {
					if _, ok := s.stopwords[term]; !ok {
						doc.Doc.Tokens = append(
							doc.Doc.Tokens,
							&xsearchpb.Token{Term: term},
						)
					}
				}
				if err := s.indexer.IndexDoc(doc.Doc); err == nil {
					log.Infof("Type=IndexDoc\tDocID=%s", doc.Doc.ID)
				}
			}
		case xsearchpb.Command_REMOVE:
			if docID, ok := cmd.DocOneof.(*xsearchpb.Command_DocID); ok {
				s.indexer.RemoveDoc(docID.DocID)
				log.Infof("Type=RemoveDoc\tDocID=%s", docID.DocID)
			}
		}
	}
	log.Trace(2, "indexLoop end")
}

func (s *shard) close() {
	s.cancel()
	s.wg.Wait()
	s.indexer.Close()
}
