package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matti/sukka/pkg/websocket"
)

func Run() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.Group("/sukka", gin.BasicAuth(gin.Accounts{
		"foo": "bar",
	})).GET("/ws", func(c *gin.Context) {
		websocket.Handler(c)
	})

	r.Run()
}
