package types

type Posting struct {
	next, prev *Posting

	list *PostingList

	DocID string
	Freq  uint32
}

func (p *Posting) Next() *Posting {
	if n := p.next; p.list != nil && n != &p.list.root {
		return n
	}
	return nil
}

func (p *Posting) Prev() *Posting {
	if n := p.prev; p.list != nil && n != &p.list.root {
		return n
	}
	return nil
}

type PostingList struct {
	root Posting
	len  int
}

func (l *PostingList) Init() *PostingList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

func NewPostingList() *PostingList { return new(PostingList).Init() }

func (l *PostingList) Len() int { return l.len }

func (l *PostingList) Front() *Posting {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

func (l *PostingList) Back() *Posting {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *PostingList) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *PostingList) insert(p, at *Posting) {
	n := at.next
	at.next = p
	p.prev = at
	p.next = n
	n.prev = p
	p.list = l
	l.len++
}

func (l *PostingList) remove(p *Posting) {
	p.prev.next = p.next
	p.next.prev = p.prev
	p.next = nil // avoid memory leaks
	p.prev = nil // avoid memory leaks
	p.list = nil
	l.len--
}

func (l *PostingList) Remove(p *Posting) {
	if p.list == l {
		l.remove(p)
	}
}

func (l *PostingList) PushFront(p *Posting) {
	l.lazyInit()
	l.insert(p, &l.root)
}

func (l *PostingList) PushBack(p *Posting) {
	l.lazyInit()
	l.insert(p, l.root.prev)
}

type Iface struct {
	Type uintptr
	Data uintptr
}

type ReqHeader struct {
	TraceID string
}
