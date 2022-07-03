package socks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/armon/go-socks5"
)

func MustParseCIDR(s string) net.IPNet {
	if _, ipnet, err := net.ParseCIDR(s); err != nil {
		panic(err)
	} else {
		return *ipnet
	}
}

var DefaultFilteredNetworks = []net.IPNet{
	MustParseCIDR("10.0.0.0/8"),         // RFC1918
	MustParseCIDR("172.16.0.0/12"),      // private
	MustParseCIDR("192.168.0.0/16"),     // private
	MustParseCIDR("127.0.0.0/8"),        // RFC5735
	MustParseCIDR("0.0.0.0/8"),          // RFC1122 Section 3.2.1.3
	MustParseCIDR("169.254.0.0/16"),     // RFC3927
	MustParseCIDR("192.0.0.0/24"),       // RFC 5736
	MustParseCIDR("192.0.2.0/24"),       // RFC 5737
	MustParseCIDR("198.51.100.0/24"),    // Assigned as TEST-NET-2
	MustParseCIDR("203.0.113.0/24"),     // Assigned as TEST-NET-3
	MustParseCIDR("192.88.99.0/24"),     // RFC 3068
	MustParseCIDR("192.18.0.0/15"),      // RFC 2544
	MustParseCIDR("224.0.0.0/4"),        // RFC 3171
	MustParseCIDR("240.0.0.0/4"),        // RFC 1112
	MustParseCIDR("255.255.255.255/32"), // RFC 919 Section 7
	MustParseCIDR("100.64.0.0/10"),      // RFC 6598
	MustParseCIDR("::/128"),             // RFC 4291: Unspecified Address
	MustParseCIDR("::1/128"),            // RFC 4291: Loopback Address
	MustParseCIDR("100::/64"),           // RFC 6666: Discard Address Block
	MustParseCIDR("2001::/23"),          // RFC 2928: IETF Protocol Assignments
	MustParseCIDR("2001:2::/48"),        // RFC 5180: Benchmarking
	MustParseCIDR("2001:db8::/32"),      // RFC 3849: Documentation
	MustParseCIDR("2001::/32"),          // RFC 4380: TEREDO
	MustParseCIDR("fc00::/7"),           // RFC 4193: Unique-Local
	MustParseCIDR("fe80::/10"),          // RFC 4291: Section 2.5.6 Link-Scoped Unicast
	MustParseCIDR("ff00::/8"),           // RFC 4291: Section 2.7
	MustParseCIDR("2002::/16"),          // RFC 7526: 6to4 anycast prefix deprecated
}

func Serve(conn net.Conn) {
	dialer := func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" && network != "tcp4" && network != "tcp6" {
			return nil, fmt.Errorf("unsupported network %s", network)
		}

		tcpAddr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			return nil, err
		}

		for _, ipnet := range DefaultFilteredNetworks {
			if ipnet.Contains(tcpAddr.IP) {
				log.Println("contains", ipnet, tcpAddr.IP)
				return nil, fmt.Errorf("blocked: IP %s is in network %s", tcpAddr.IP, ipnet)
			}
		}

		d := net.Dialer{
			Timeout: 30 * time.Second,
		}

		return d.Dial(network, address)
	}
	socks, err := socks5.New(&socks5.Config{
		Dial: dialer,
	})
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	err = socks.ServeConn(conn)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		if err.Error() == "Failed to handle request: io: read/write on closed pipe" {
			return
		}
		if err.Error() == "Failed to handle request: stream closed" {
			return
		}
		log.Println("socks serve", err)
		return
	}
}
