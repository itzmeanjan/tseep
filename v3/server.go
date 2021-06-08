package v3

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/itzmeanjan/tseep/op"
	"github.com/xtaci/gaio"
	pool "gopkg.in/thejerf/gomempool.v1"
)

type Server struct {
	Addr         string
	Listener     net.Listener
	WatcherCount uint
	Watchers     map[uint]*watcher
	KV           map[op.Key]op.Value
	KVLock       *sync.RWMutex
	Pool         *pool.Pool
}

type watcher struct {
	eventPool      *gaio.Watcher
	inProgressRead map[net.Conn]*readingState
	lock           *sync.RWMutex
}

type readingState struct {
	allocator    pool.Allocator
	envelopeRead bool
	opcode       op.OP
}

func New(ctx context.Context, proto string, addr string, watcherCount uint) (*Server, error) {
	lis, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	srv := Server{
		KV:           make(map[op.Key]op.Value),
		KVLock:       &sync.RWMutex{},
		Addr:         lis.Addr().String(),
		Listener:     lis,
		Pool:         pool.New(1<<16, 1<<24, 1<<4),
		WatcherCount: watcherCount,
		Watchers:     make(map[uint]*watcher),
	}

	watcherChan := make(chan struct{}, watcherCount)
	var i uint
	for ; i < srv.WatcherCount; i++ {
		w, err := gaio.NewWatcher()
		if err != nil {
			return nil, err
		}

		srv.Watchers[i] = &watcher{
			eventPool:      w,
			inProgressRead: make(map[net.Conn]*readingState),
			lock:           &sync.RWMutex{},
		}
		func(id uint) {
			go srv.Watch(ctx, id, watcherChan)
		}(i)
	}

	lisChan := make(chan struct{})
	go srv.Listen(ctx, lisChan)
	<-lisChan

	running := 0
	for range watcherChan {
		running++
		if running >= int(watcherCount) {
			break
		}
	}

	return &srv, nil
}

func (s *Server) Listen(ctx context.Context, done chan struct{}) {
	close(done)
	defer func() {
		if err := s.Listener.Close(); err != nil {
			log.Printf("Failed to close listener : %s\n", err.Error())
		}
	}()

	var nextWatcher uint
	for {
		select {
		case <-ctx.Done():
			return

		default:
			conn, err := s.Listener.Accept()
			if err != nil {
				log.Printf("Server not listening : %s\n", err.Error())
				return
			}

			watcher := s.Watchers[nextWatcher]
			allocator := s.Pool.GetNewAllocator()
			watcher.lock.Lock()
			watcher.inProgressRead[conn] = &readingState{allocator: allocator}
			watcher.lock.Unlock()

			if err := watcher.eventPool.Read(ctx, conn, allocator.Allocate(3)); err != nil {
				return
			}

			nextWatcher = (nextWatcher + 1) * s.WatcherCount

		}
	}
}

func (s *Server) Watch(ctx context.Context, id uint, done chan struct{}) {
	done <- struct{}{}
	watcher := s.Watchers[id]
	defer func() {
		if err := watcher.eventPool.Close(); err != nil {
			log.Printf("Failed to close watcher : %s [ %d ]\n", err.Error(), id)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		default:
			results, err := watcher.eventPool.WaitIO()
			if err != nil {
				return
			}

			for _, result := range results {
				switch result.Operation {
				case gaio.OpRead:
					if err := s.handleRead(ctx, result, watcher); err != nil {
						watcher.lock.Lock()
						delete(watcher.inProgressRead, result.Conn)
						watcher.lock.Unlock()

						watcher.eventPool.Free(result.Conn)
					}

				case gaio.OpWrite:
					if err := s.handleWrite(ctx, result, watcher); err != nil {
						watcher.lock.Lock()
						delete(watcher.inProgressRead, result.Conn)
						watcher.lock.Unlock()

						watcher.eventPool.Free(result.Conn)
					}

				}
			}
		}
	}
}

func (s *Server) handleRead(ctx context.Context, result gaio.OpResult, watcher *watcher) error {
	if result.Error != nil {
		return result.Error
	}

	if result.Size == 0 {
		return errors.New("empty read")
	}

	watcher.lock.RLock()
	defer watcher.lock.RUnlock()
	v := watcher.inProgressRead[result.Conn]
	if !v.envelopeRead {
		r := bytes.NewReader(result.Buffer[:])
		opcode, bodyLen, err := op.ReadEnvelope(r)
		if err != nil {
			return err
		}

		v.opcode = opcode
		v.envelopeRead = true

		v.allocator.Return()
		return watcher.eventPool.Read(ctx, result.Conn, v.allocator.Allocate(uint64(bodyLen)))
	}

	r := bytes.NewReader(result.Buffer[:])
	w := new(bytes.Buffer)

	switch v.opcode {
	case op.READ:
		rReq := new(op.ReadRequest)
		if _, err := rReq.ReadFrom(r); err != nil {
			return err
		}

		s.KVLock.RLock()
		val, ok := s.KV[*rReq.Key]
		if !ok {
			val = op.Value([]byte(""))
		}
		s.KVLock.RUnlock()

		if _, err := val.WriteTo(w); err != nil {
			return err
		}

	case op.WRITE:
		wReq := new(op.WriteRequest)
		if _, err := wReq.ReadFrom(r); err != nil {
			return err
		}

		s.KVLock.Lock()
		s.KV[*wReq.Key] = *wReq.Value
		s.KVLock.Unlock()

		if _, err := wReq.Value.WriteTo(w); err != nil {
			return err
		}

	default:
		return errors.New("bad opcode")

	}

	v.allocator.Return()
	return watcher.eventPool.Write(ctx, result.Conn, w.Bytes())
}

func (s *Server) handleWrite(ctx context.Context, result gaio.OpResult, watcher *watcher) error {
	if result.Error != nil {
		return result.Error
	}

	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	v := watcher.inProgressRead[result.Conn]
	v.envelopeRead = false

	return watcher.eventPool.Read(ctx, result.Conn, v.allocator.Allocate(3))
}
