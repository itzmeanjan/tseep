package v3

import (
	"context"
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
