package indexer

import (
	"encoding/binary"
	"math"
	"os"
	"syscall"
	"testing"
	"unsafe"

	"github.com/tddhit/tools/log"
)

const (
	dictPath      = "../dict/dictionary.txt"
	stopwordsPath = "../dict/stopwords.txt"
)

/*
func TestIndexer(t *testing.T) {
	title := "绿箭侠是一部由剧作家/制片人 Greg Berlanti、Marc Guggenheim和Andrew Kreisberg创作的电视连续剧。它基于DC漫画角色绿箭侠，一个由Mort Weisinger和George Papp创作的装扮奇特的犯罪打击战士。"
	idx := New(dictPath)
	util.InitStopwords(stopwordsPath)
	idx.IndexDocument(&types.Document{Title: title})
	for k, l := range idx.Dict {
		fmt.Println(k)
		var next *list.Element
		for e := l.Front(); e != nil; e = next {
			fmt.Println(*e.Value.(*types.Posting))
			next = e.Next()
		}
	}
}
*/

func TestMmap(t *testing.T) {
	invertFile, err := os.Open("invert.bin")
	if err != nil {
		log.Panic(err)
	}
	invertRef, err := syscall.Mmap(int(invertFile.Fd()), 0, MaxMapSize,
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Panic(err)
	}
	if _, _, err := syscall.Syscall(syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&invertRef[0])), uintptr(len(invertRef)),
		uintptr(syscall.MADV_RANDOM)); err != 0 {

		log.Panic(err)
	}
	fi, err := invertFile.Stat()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Start")
	var (
		loc   = 0
		count uint64
	)
	for {
		if loc >= int(fi.Size()) {
			break
		}
		count = binary.LittleEndian.Uint64(invertRef[loc : loc+8 : loc+8])
		loc += 8
		for i := uint64(0); i < count; i++ {
			binary.LittleEndian.Uint64(invertRef[loc : loc+8 : loc+8])
			loc += 8
			bits := binary.LittleEndian.Uint32(invertRef[loc : loc+4 : loc+4])
			math.Float32frombits(bits)
			loc += 4
		}
	}
	log.Info("End")
}

func TestMem(t *testing.T) {
	m := make(map[uint64]int, 500000)
	log.Info("start")
	for i := uint64(0); i < 380000; i++ {
		m[i] = 1
	}
	log.Info("end")
}

func TestReadFile(t *testing.T) {
	invertFile, err := os.Open("invert.bin")
	if err != nil {
		log.Panic(err)
	}
	log.Info("Start")
	var (
		loc   int64 = 0
		count uint64
	)
	for {
		if loc == 333754368 {
			break
		}
		buf := make([]byte, 8)
		n, err := invertFile.ReadAt(buf, loc)
		if err != nil || n != cap(buf) {
			log.Fatal(err)
		}
		count = binary.LittleEndian.Uint64(buf)
		loc += 8
		for i := uint64(0); i < count; i++ {
			var (
				docID uint64
				bm25  float32
			)
			buf := make([]byte, 8)
			n, err := invertFile.ReadAt(buf, loc)
			if err != nil || n != cap(buf) {
				log.Fatal(err)
			}
			docID = binary.LittleEndian.Uint64(buf)
			docID = docID
			loc += 8
			buf = make([]byte, 4)
			n, err = invertFile.ReadAt(buf, loc)
			if err != nil || n != cap(buf) {
				log.Fatal(err)
			}
			bits := binary.LittleEndian.Uint32(buf)
			bm25 = math.Float32frombits(bits)
			bm25 = bm25
			loc += 4
		}
	}
	log.Info("End")
}
