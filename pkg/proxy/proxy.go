package proxy

import (
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/matti/betterio"
)

type Stats struct {
	inflight      uint64
	bytesSent     int64
	bytesReceived int64
}

var stats Stats

func StatsPrinter() {
	for {
		log.Println("inflight", stats.inflight, "tx", stats.bytesSent/1024/1024, "MB", "rx", stats.bytesReceived/1024/1024, "MB")
		time.Sleep(1 * time.Second)
	}
}

func Run(upstreams chan net.Conn) {
	go StatsPrinter()

	ln, err := net.Listen("tcp", "127.0.0.1:1080")
	if err != nil {
		log.Panic("listen err", err)
	}

	for {
		upstream := <-upstreams
		downstream, err := ln.Accept()
		if err != nil {
			// temporary error
			log.Panicln("accept", err)
		}
		atomic.AddUint64(&stats.inflight, 1)
		go func() {
			defer downstream.Close()
			defer upstream.Close()
			bytesUp, bytesDown := betterio.CopyBidirUntilCloseAndReturnBytesWritten(downstream, upstream)

			atomic.AddInt64(&stats.bytesSent, bytesUp)
			atomic.AddInt64(&stats.bytesReceived, bytesDown)

			for {
				now := atomic.LoadUint64(&stats.inflight)
				if atomic.CompareAndSwapUint64(&stats.inflight, now, now-1) {
					break
				}
			}
		}()
	}
}
