package v2

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
	Addr           string
	Listener       net.Listener
	Watcher        *gaio.Watcher
	KV             map[op.Key]op.Value
	KVLock         *sync.RWMutex
	InProgressRead map[net.Conn]*readBuffer
	ReadLock       *sync.RWMutex
	Pool           *pool.Pool
}

type readBuffer struct {
	allocator    pool.Allocator
	envelopeRead bool
	opcode       op.OP
}

func New(ctx context.Context, proto string, addr string) (*Server, error) {
	lis, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	watcher, err := gaio.NewWatcher()
	if err != nil {
		return nil, err
	}

	srv := Server{
		KV:             make(map[op.Key]op.Value),
		KVLock:         &sync.RWMutex{},
		InProgressRead: make(map[net.Conn]*readBuffer),
		ReadLock:       &sync.RWMutex{},
		Listener:       lis,
		Addr:           lis.Addr().String(),
		Watcher:        watcher,
		Pool:           pool.New(1<<16, 1<<24, 1<<4),
	}

	lisChan := make(chan struct{})
	watcherChan := make(chan struct{})
	go srv.Listen(ctx, lisChan)
	go srv.Watch(ctx, watcherChan)
	<-lisChan
	<-watcherChan

	return &srv, nil
}

func (s *Server) Listen(ctx context.Context, done chan struct{}) {
	close(done)
	defer func() {
		if err := s.Listener.Close(); err != nil {
			log.Printf("Failed to close listener : %s\n", err.Error())
		}
	}()

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

			allocator := s.Pool.GetNewAllocator()
			s.ReadLock.Lock()
			s.InProgressRead[conn] = &readBuffer{allocator: allocator}
			s.ReadLock.Unlock()

			if err := s.Watcher.Read(ctx, conn, allocator.Allocate(3)); err != nil {
				return
			}
		}
	}
}

func (s *Server) Watch(ctx context.Context, done chan struct{}) {
	close(done)
	defer func() {
		if err := s.Watcher.Close(); err != nil {
			log.Printf("Failed to close watcher : %s\n", err.Error())
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		default:
			results, err := s.Watcher.WaitIO()
			if err != nil {
				return
			}

			for _, result := range results {
				switch result.Operation {
				case gaio.OpRead:
					if err := s.handleRead(ctx, result); err != nil {
						s.ReadLock.Lock()
						delete(s.InProgressRead, result.Conn)
						s.ReadLock.Unlock()

						s.Watcher.Free(result.Conn)
					}

				case gaio.OpWrite:
					if err := s.handleWrite(ctx, result); err != nil {
						s.ReadLock.Lock()
						delete(s.InProgressRead, result.Conn)
						s.ReadLock.Unlock()

						s.Watcher.Free(result.Conn)
					}

				}
			}
		}
	}
}

func (s *Server) handleRead(ctx context.Context, result gaio.OpResult) error {
	if result.Error != nil {
		return result.Error
	}

	if result.Size == 0 {
		return errors.New("empty read")
	}

	s.ReadLock.RLock()
	defer s.ReadLock.RUnlock()
	v := s.InProgressRead[result.Conn]
	if !v.envelopeRead {
		r := bytes.NewReader(result.Buffer[:])
		opcode, bodyLen, err := op.ReadEnvelope(r)
		if err != nil {
			return err
		}

		v.opcode = opcode
		v.envelopeRead = true

		v.allocator.Return()
		return s.Watcher.Read(ctx, result.Conn, v.allocator.Allocate(uint64(bodyLen)))
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
	return s.Watcher.Write(ctx, result.Conn, w.Bytes())
}

func (s *Server) handleWrite(ctx context.Context, result gaio.OpResult) error {
	if result.Error != nil {
		return result.Error
	}

	s.ReadLock.RLock()
	defer s.ReadLock.RUnlock()

	v := s.InProgressRead[result.Conn]
	v.envelopeRead = false

	return s.Watcher.Read(ctx, result.Conn, v.allocator.Allocate(3))
}
