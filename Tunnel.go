package tunnel

import (
	"fmt"
	"io"
	"log"
	"net/url"
)

const BUFF_SIZE = 0xffff

type endpoint interface {
	ListenAndServe() error
	Dial() (io.ReadWriteCloser, error)
	Close() error
}

type void struct{}

type Tunnel struct {
	lEndpoint endpoint
	rEndpoint endpoint
	cID       int
	channels  map[int]chan void
}

func NewTunnel(lu, ru *url.URL) (t *Tunnel, err error) {
	switch lu.Scheme {
	case "http":
		e, err := newHttpEndpoint(lu, t)
		if err != nil {
			return nil, err
		}
		t.lEndpoint = e
	case "tcp":
		e, err := newTcpEndpoint(lu, t)
		if err != nil {
			return nil, err
		}
		t.lEndpoint = e
	default:
		return nil, fmt.Errorf("laddr scheme is not supported")
	}
	switch ru.Scheme {
	case "http":
		e, err := newHttpEndpoint(lu, t)
		if err != nil {
			return nil, err
		}
		t.rEndpoint = e
	case "tcp":
		e, err := newTcpEndpoint(lu, t)
		if err != nil {
			return nil, err
		}
		t.rEndpoint = e
	default:
		return nil, fmt.Errorf("raddr scheme is not supported")
	}
	return
}

func (t *Tunnel) Start() error {
	return t.lEndpoint.ListenAndServe()
}

func (t *Tunnel) newConnection(lConn io.ReadWriteCloser) {
	rConn, err := t.rEndpoint.Dial()
	if err != nil {
		log.Printf("New Connection: %s\n", err)
		return
	}
	defer lConn.Close()
	defer rConn.Close()
	t.channels[t.cID] = make(chan void)
	go t.pipe(lConn, rConn, t.cID)
	go t.pipe(rConn, lConn, t.cID)
	t.cID++
	<-t.channels[t.cID]
}

func (t *Tunnel) Errorf(id int, fmt string, v ...any) {
	log.Printf(fmt, v...)
	t.channels[id] <- void{}
}

func (t *Tunnel) pipe(src, dst io.ReadWriter, id int) {
	buff := make([]byte, BUFF_SIZE)
	for {
		n, err := src.Read(buff)
		if err != nil {
			t.Errorf(id, "Read failed '%s'\n", err)
			return
		}
		b := buff[:n]
		n, err = dst.Write(b)
		if err != nil {
			t.Errorf(id, "Write failed '%s'\n", err)
			return
		}
	}
}

func (t *Tunnel) Close() error {
	t.lEndpoint.Close()
	t.rEndpoint.Close()
	return nil
}
