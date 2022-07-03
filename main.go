package main

import (
	"net"
	"os"

	"github.com/matti/sukka/pkg/httpserver"
	"github.com/matti/sukka/pkg/proxy"
	"github.com/matti/sukka/pkg/sockspipe"
	"github.com/matti/sukka/pkg/websocket"
)

func main() {
	switch os.Args[1] {
	case "server":
		httpserver.Run()
	case "client", "proxy":
		// size of 1 = 3 when client connected: currently using, next one and the 1 that blocks
		// and size of 3 = 5 --> 1=3 and 2 = 5
		upstreams := make(chan net.Conn, 4)

		switch os.Args[1] {
		case "client":
			go websocket.Dialer(upstreams, os.Args[2])
		case "proxy":
			go sockspipe.Piper(upstreams)
		}

		proxy.Run(upstreams)
	}
}
