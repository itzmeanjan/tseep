package v1

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/itzmeanjan/tseep/op"
)

type Server struct {
	Addr     string
	Listener net.Listener
	KV       map[op.Key]op.Value
	Lock     *sync.RWMutex
}

func New(ctx context.Context, proto string, addr string) (*Server, error) {
	lis, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	srv := Server{
		KV:       make(map[op.Key]op.Value),
		Lock:     &sync.RWMutex{},
		Listener: lis,
		Addr:     lis.Addr().String(),
	}

	done := make(chan struct{})
	go srv.Start(ctx, done)
	<-done

	return &srv, nil
}

func (s *Server) Start(ctx context.Context, done chan struct{}) {
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

			func(conn net.Conn) {
				go s.handleConnection(ctx, conn)
			}(conn)
		}
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection : %s\n", err.Error())
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		default:
			opcode, _, err := op.ReadEnvelope(conn)
			if err != nil {
				return
			}

			switch opcode {
			case op.READ:
				rReq := new(op.ReadRequest)
				if _, err := rReq.ReadFrom(conn); err != nil {
					return
				}

				s.Lock.RLock()
				val, ok := s.KV[*rReq.Key]
				if !ok {
					val = op.Value([]byte(""))
				}
				s.Lock.RUnlock()

				if _, err := val.WriteTo(conn); err != nil {
					return
				}

			case op.WRITE:
				wReq := new(op.WriteRequest)
				if _, err := wReq.ReadFrom(conn); err != nil {
					return
				}

				s.Lock.Lock()
				s.KV[*wReq.Key] = *wReq.Value
				s.Lock.Unlock()

				if _, err := wReq.Value.WriteTo(conn); err != nil {
					return
				}

			default:
				return

			}

		}
	}

}
