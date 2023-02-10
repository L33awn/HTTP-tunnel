package tunnel

import (
	"fmt"
	"io"
	"net"
	"net/url"
)

type TcpEndpoint struct {
	address  string
	listener *net.Listener
	tunnel   *Tunnel
}

func newTcpEndpoint(u *url.URL, t *Tunnel) (te *TcpEndpoint, err error) {
	return &TcpEndpoint{
		address: fmt.Sprintf("%s:%s", u.Host, u.Port()),
		tunnel:  t,
	}, nil
}

func (te *TcpEndpoint) ListenAndServe() (err error) {
	l, err := net.Listen(
		"tcp",
		te.address,
	)
	if err != nil {
		return
	}
	te.listener = &l
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go te.tunnel.newConnection(conn)
	}
}

func (te *TcpEndpoint) Dial() (io.ReadWriteCloser, error) {
	return net.Dial("tcp", fmt.Sprintf("%s:%d", te.address))
}

func (te *TcpEndpoint) Close() (err error) {
	return (*te.listener).Close()
}
