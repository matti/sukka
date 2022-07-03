package muxer

import (
	"log"
	"net"
	"time"

	"github.com/hashicorp/yamux"
)

func Client(upstreams chan net.Conn, conn net.Conn) {
	errCh := make(chan error)

	session, err := yamux.Client(conn, nil)
	if err != nil {
		errCh <- err
	}

	go func() {
		for {
			conn, err := session.Open()
			if err != nil {
				errCh <- err
				return
			}
			upstreams <- conn
		}
	}()

	go func() {
		for {
			_, err := session.Ping()
			if err != nil {
				errCh <- err
			}

			time.Sleep(1 * time.Second)
		}
	}()

	err = <-errCh
	log.Println("err", err)
	session.Close()
}
