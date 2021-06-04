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
)

type Server struct {
	Addr           string
	Listener       net.Listener
	Watcher        *gaio.Watcher
	KV             map[op.Key]op.Value
	KVLock         *sync.RWMutex
	InProgressRead map[net.Conn]*readBuffer
	ReadLock       *sync.RWMutex
}

type readBuffer struct {
	buf          []byte
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

			envelopeBuf := make([]byte, 3)
			s.ReadLock.Lock()
			s.InProgressRead[conn] = &readBuffer{buf: envelopeBuf}
			s.ReadLock.Unlock()

			if err := s.Watcher.Read(ctx, conn, envelopeBuf); err != nil {
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

		return s.Watcher.Read(ctx, result.Conn, make([]byte, bodyLen))
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

	return s.Watcher.Read(ctx, result.Conn, v.buf)
}