package streams

import (
	"sync"
)

// 如回调函数返回false，则删除注册的回调
type HandlerFunc func(data interface{}) bool

// 需要有个结构，不能只用一个回调函数，因为func不允许比较
type Handler struct {
	handle HandlerFunc
	quitCh chan bool
}

type Stream struct {
	Name     string
	C        chan interface{}
	MaxLen   int
	mutex    *sync.Mutex
	handlers []*Handler
}

func NewStream(name string, maxlen int) *Stream {
	stream := &Stream{
		Name:     name,
		MaxLen:   maxlen,
		C:        make(chan interface{}, maxlen),
		handlers: []*Handler{},
		mutex:    &sync.Mutex{},
	}
	return stream
}

func (s *Stream) Pub(data interface{}) {
	if len(s.C) < s.MaxLen {
		s.C <- data
	}
}

func (s *Stream) Sub(fun HandlerFunc) <-chan bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	h := &Handler{handle: fun, quitCh: make(chan bool)}
	s.handlers = append(s.handlers, h)

	// 结束信号
	return h.quitCh
}

func (s *Stream) Run() {
	for {
		data := <-s.C

		s.mutex.Lock()
		for _, handler := range s.handlers {
			if !handler.handle(data) {
				s.removeHandlerFunc(handler)
			}
		}
		s.mutex.Unlock()
	}
}

func (s *Stream) removeHandlerFunc(handler *Handler) {
	handlers := []*Handler{}
	for _, h := range s.handlers {
		if h != handler {
			handlers = append(handlers, h)
		}
	}
	handler.quitCh <- true
	s.handlers = handlers
}
