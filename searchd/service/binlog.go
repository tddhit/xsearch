package service

import "github.com/tddhit/tools/mmap"

type binLog struct {
	dir  string
	name string
	file *mmap.MmapFile
}

func newBinLog(name string) (*binlog, error) {
	l := &binLog{
		name: name,
	}
	return l, nil
}

func (l *binLog) writeLoop() {
	for {
		select {
		case entry := <-l.entryC:
			l.writeOne(entry)
		case l.exitC:
			goto exit
		}
	}
exit:
}

func (l *binlog) writeOne(entry *pb.LogEntry) error {
	buf, err := proto.Marshal(entry)
	if err != nil {
		log.Error(err)
		return err
	}
	len := uint32(len(buf))
	if len > maxLogSize {
		return fmt.Errorf("invalid msg:size(%d)>maxSize(%d)", len, maxLogSize)
	}
	offset := atomic.LoadInt64(&l.writePos)
	if err := l.file.PutUint32At(offset, len); err != nil {
		return err
	}
	offset += int64(len)
	atomic.StoreInt64(&l.writePos, offset)
	return nil
}

func (l *binLog) writeLog(data []byte) error {
	l.RLock()
	defer l.RUnlock()

	l.writeC <- &pb.LogEntry{
		Action: pb.ADD,
		Data:   data,
	}
	return <-t.writeRspC
}
