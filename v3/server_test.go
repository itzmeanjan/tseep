package v3_test

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/itzmeanjan/tseep/op"
	v3 "github.com/itzmeanjan/tseep/v3"
)

func TestServerV3(t *testing.T) {
	proto := "tcp"
	addr := "127.0.0.1:0"
	watcherCount := 4
	ctx, cancel := context.WithCancel(context.Background())
	server, err := v3.New(ctx, proto, addr, uint(watcherCount))
	if err != nil {
		t.Fatalf("Failed to start TCP server : %s\n", err.Error())
	}

	testClientFlow(t, ctx, proto, server.Addr)
	cancel()
}

func testClientFlow(t *testing.T, ctx context.Context, proto string, addr string) {
	conn, err := net.Dial(proto, addr)
	if err != nil {
		t.Fatalf("Failed to dial TCP server : %s\n", err.Error())
	}
	defer func() {
		conn.Close()
	}()

	key := op.Key("hello")
	rReq := op.ReadRequest{Key: &key}
	w := new(bytes.Buffer)
	if _, err := rReq.WriteEnvelope(w); err != nil {
		t.Fatalf("Failed to write request envelope : %s\n", err.Error())
	}

	if _, err := conn.Write(w.Bytes()); err != nil {
		t.Fatalf("Failed to write request envelope : %s\n", err.Error())
	}

	w.Reset()

	if _, err := rReq.WriteTo(w); err != nil {
		t.Fatalf("Failed to write request body : %s\n", err.Error())
	}

	if _, err := conn.Write(w.Bytes()); err != nil {
		t.Fatalf("Failed to write request body : %s\n", err.Error())
	}

	w.Reset()

	resp := new(op.Value)
	if _, err := resp.ReadFrom(conn); err != nil {
		t.Fatalf("Failed to read response : %s\n", err.Error())
	}

	if !bytes.Equal(*resp, []byte("")) {
		t.Fatalf("Expected to receive empty response\n")
	}

	wVal := op.Value("world")
	wReq := op.WriteRequest{Key: &key, Value: &wVal}
	if _, err := wReq.WriteEnvelope(w); err != nil {
		t.Fatalf("Failed to write request envelope : %s\n", err.Error())
	}

	if _, err := conn.Write(w.Bytes()); err != nil {
		t.Fatalf("Failed to write request envelope : %s\n", err.Error())
	}

	w.Reset()

	if _, err := wReq.WriteTo(w); err != nil {
		t.Fatalf("Failed to write request body : %s\n", err.Error())
	}

	if _, err := conn.Write(w.Bytes()); err != nil {
		t.Fatalf("Failed to write request body : %s\n", err.Error())
	}

	w.Reset()

	if _, err := resp.ReadFrom(conn); err != nil {
		t.Fatalf("Failed to read response : %s\n", err.Error())
	}

	if !bytes.Equal(*resp, wVal) {
		t.Fatalf("Expected to receive `%s`, received `%s`\n", wVal, *resp)
	}

	if _, err := rReq.WriteEnvelope(w); err != nil {
		t.Fatalf("Failed to write request envelope : %s\n", err.Error())
	}

	if _, err := conn.Write(w.Bytes()); err != nil {
		t.Fatalf("Failed to write request envelope : %s\n", err.Error())
	}

	w.Reset()

	if _, err := rReq.WriteTo(w); err != nil {
		t.Fatalf("Failed to write request body : %s\n", err.Error())
	}

	if _, err := conn.Write(w.Bytes()); err != nil {
		t.Fatalf("Failed to write request body : %s\n", err.Error())
	}

	w.Reset()

	if _, err := resp.ReadFrom(conn); err != nil {
		t.Fatalf("Failed to read response : %s\n", err.Error())
	}

	if !bytes.Equal(wVal, *resp) {
		t.Fatalf("Expected to receive `%s`, received `%s`\n", wVal, *resp)
	}
}
