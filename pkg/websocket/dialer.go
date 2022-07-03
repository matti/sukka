package websocket

import (
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/matti/sukka/pkg/muxer"
)

func Dialer(upstreams chan net.Conn, serverUrl string) {
	for {
		Dial(upstreams, serverUrl)
		time.Sleep(1 * time.Second)
	}
}

func Dial(upstreams chan net.Conn, serverUrl string) {
	dialer := &websocket.Dialer{
		HandshakeTimeout: 3 * time.Second,
	}

	u, err := url.Parse(serverUrl)
	if err != nil {
		log.Panic("invalid url", err)
	}
	dialUrl := strings.Join([]string{
		u.Scheme,
		"://",
		u.Host,
		u.Path,
	}, "")

	log.Println(dialUrl)
	headers := http.Header{
		"Authorization": []string{
			"Basic " + base64.StdEncoding.EncodeToString([]byte(u.User.String())),
		},
	}

	wsConn, res, err := dialer.Dial(dialUrl, headers)
	if err != nil {
		log.Println("dial", err)
		return
	}

	if res.StatusCode != 101 {
		log.Println("status not 101", res.StatusCode)
		return
	}

	defer wsConn.Close()

	conn := wsConn.UnderlyingConn()

	muxer.Client(upstreams, conn)
	conn.Close()
}
