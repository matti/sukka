package websocket

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/matti/sukka/pkg/socks"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Second * 3,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handler(c *gin.Context) {
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws", "err", err)
		return
	}
	defer wsConn.Close()

	conn := wsConn.UnderlyingConn()
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.Panic("server", err)
	}

	// keep websocket connection open until client disconnects
	for {
		// accepts initially as many times as client's downstream channel size is regardless of the active connections to speed up
		downstream, err := session.Accept()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// log.Println("client disconnected")
				return
			}
			log.Panic("session accept", err)
		}

		go func() {
			socks.Serve(downstream)
			downstream.Close()
		}()
	}
}
