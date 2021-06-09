// +build stress

package v3_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/itzmeanjan/tseep/op"
	v3 "github.com/itzmeanjan/tseep/v3"
)

func TestServerV3_Stress_1k(t *testing.T) {
	concurrentConnectionTest(t, 1<<10)
}

func TestServerV3_Stress_2k(t *testing.T) {
	concurrentConnectionTest(t, 1<<11)
}

func TestServerV3_Stress_4k(t *testing.T) {
	concurrentConnectionTest(t, 1<<12)
}

func TestServerV3_Stress_8k(t *testing.T) {
	concurrentConnectionTest(t, 1<<13)
}

func concurrentConnectionTest(t *testing.T, clientCount int) {
	proto := "tcp"
	addr := "127.0.0.1:0"
	watcherCount := uint(4)
	ctx, cancel := context.WithCancel(context.Background())
	server, err := v3.New(ctx, proto, addr, watcherCount)
	if err != nil {
		t.Fatalf("Failed to start TCP server : %s\n", err.Error())
	}
	defer func() {
		server.Listener.Close()
	}()

	report := make(chan struct{}, clientCount)
	for i := 0; i < clientCount; i++ {
		func(idx int) {

			go func() {
				defer func() {
					report <- struct{}{}
				}()

				conn, err := net.Dial(proto, server.Addr)
				if err != nil {
					t.Logf("Failed to dial : %s\n", err.Error())
					return
				}
				defer func() {
					conn.Close()
					report <- struct{}{}
				}()
				w := new(bytes.Buffer)

				key := op.Key(fmt.Sprintf("%255d", idx))
				rReq := op.ReadRequest{Key: &key}

				if _, err := rReq.WriteEnvelope(w); err != nil {
					return
				}
				if _, err := conn.Write(w.Bytes()); err != nil {
					return
				}
				w.Reset()

				if _, err := rReq.WriteTo(w); err != nil {
					return
				}
				if _, err := conn.Write(w.Bytes()); err != nil {
					return
				}
				w.Reset()

				resp := new(op.Value)
				if _, err := resp.ReadFrom(conn); err != nil {
					return
				}

				wVal := op.Value(fmt.Sprintf("%255d", idx))
				wReq := op.WriteRequest{Key: &key, Value: &wVal}

				if _, err := wReq.WriteEnvelope(w); err != nil {
					return
				}
				if _, err := conn.Write(w.Bytes()); err != nil {
					return
				}
				w.Reset()

				if _, err := wReq.WriteTo(w); err != nil {
					return
				}
				if _, err := conn.Write(w.Bytes()); err != nil {
					return
				}
				w.Reset()

				if _, err := resp.ReadFrom(conn); err != nil {
					return
				}

				if !bytes.Equal(*resp, wVal) {
					return
				}

			}()

		}(i + 1%256)
	}

	var done int
	for range report {
		done++
		if done >= clientCount {
			break
		}
	}

	cancel()
}
