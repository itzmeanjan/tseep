package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itzmeanjan/tseep/op"
	"github.com/itzmeanjan/tseep/utils"
)

func main() {
	clientCount := utils.GetClientCount()
	clients := make([]net.Conn, 0, clientCount)
	proto := "tcp"
	addr := fmt.Sprintf("%s:%d", utils.GetAddr(), utils.GetPort())
	delay := utils.GetDelay()

	for i := 0; i < int(clientCount); i++ {
		conn, err := net.Dial(proto, addr)
		if err != nil {
			log.Printf("Failed to dial TCP server : %s\n", err.Error())
			continue
		}
		defer func() {
			conn.Close()
		}()

		clients = append(clients, conn)
	}

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGTERM, syscall.SIGINT)

OUTER:
	for {
		select {
		case <-interruptChan:
			break OUTER

		default:
			if !read_write(clients, interruptChan) {
				break OUTER
			}

			<-time.After(delay)
		}
	}

	log.Println("Graceful shutdown")
}

func read_write(clients []net.Conn, stopChan chan os.Signal) bool {
OUTER:
	for _, conn := range clients {
		select {
		case <-stopChan:
			return false

		default:
			r := rand.Intn(1 << 30)

			switch r%3 == 0 {
			case true:
				key := op.Key(fmt.Sprintf("%d", r))
				req := op.ReadRequest{Key: &key}
				if _, err := req.WriteEnvelope(conn); err != nil {
					log.Printf("Failed to write request envelope : %s\n", err.Error())
					continue OUTER
				}

				if _, err := req.WriteTo(conn); err != nil {
					log.Printf("Failed to write request body : %s\n", err.Error())
					continue OUTER
				}

				resp := new(op.Value)
				if _, err := resp.ReadFrom(conn); err != nil {
					log.Printf("Failed to read response : %s\n", err.Error())
					continue OUTER
				}

				log.Printf("Read %s => `%s` [%s]\n", key, *resp, conn.LocalAddr())

			case false:
				key := op.Key(fmt.Sprintf("%d", r))
				val := op.Value(fmt.Sprintf("%d", r))
				req := op.WriteRequest{Key: &key, Value: &val}
				if _, err := req.WriteEnvelope(conn); err != nil {
					log.Printf("Failed to write request envelope : %s\n", err.Error())
					continue OUTER
				}

				if _, err := req.WriteTo(conn); err != nil {
					log.Printf("Failed to write request body : %s\n", err.Error())
					continue OUTER
				}

				resp := new(op.Value)
				if _, err := resp.ReadFrom(conn); err != nil {
					log.Printf("Failed to read response : %s\n", err.Error())
					continue OUTER
				}

				log.Printf("Wrote %s => %s [%s]\n", key, *resp, conn.LocalAddr())

			}
		}
	}

	return true
}
