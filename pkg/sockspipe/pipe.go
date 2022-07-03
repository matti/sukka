package sockspipe

import (
	"net"

	"github.com/matti/sukka/pkg/socks"
)

func Piper(upstreams chan net.Conn) {
	for {
		theirs, ours := net.Pipe()

		// creates 3 pipes when upstream chan size == 1
		upstreams <- theirs
		go socks.Serve(ours)
	}
}
