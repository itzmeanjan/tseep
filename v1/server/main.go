package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itzmeanjan/tseep/utils"
	v1 "github.com/itzmeanjan/tseep/v1"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	srv, err := v1.New(ctx, "tcp", fmt.Sprintf("%s:%d", utils.GetAddr(), utils.GetPort()))
	if err != nil {
		log.Printf("Failed to start server : %s\n", err.Error())
		return
	}

	log.Printf("Server listening on %s\b", srv.Addr)

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGTERM, syscall.SIGINT)
	<-interruptChan

	cancel()
	<-time.After(time.Second)
}
