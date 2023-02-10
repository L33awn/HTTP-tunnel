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
	t = &Tunnel{
		channels: make(map[int]chan void),
	}
	switch lu.Scheme {
	case "http":
		e, err := newHttpEndpoint(lu, t, true)
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
		e, err := newHttpEndpoint(ru, t, false)
		if err != nil {
			return nil, err
		}
		t.rEndpoint = e
	case "tcp":
		e, err := newTcpEndpoint(ru, t)
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
	tid := t.cID
	t.channels[tid] = make(chan void)
	go t.pipe(lConn, rConn, tid)
	go t.pipe(rConn, lConn, tid)
	t.cID++
	<-t.channels[tid]
	// log.Println(lConn.Close())
	// log.Println(rConn.Close())
	fmt.Println("Connection", tid, "Closed")
}

func (t *Tunnel) Errorf(id int, fmt string, v ...any) {
	log.Printf(fmt, v...)
	t.closeConn(id)
}

func (t *Tunnel) pipe(src, dst io.ReadWriter, id int) {
	buff := make([]byte, BUFF_SIZE)
	for {
		n, err := src.Read(buff)
		if err != nil && err != io.EOF {
			t.Errorf(id, "Read failed '%s'\n", err)
			return
		}
		if n == 0 {
			t.closeConn(id)
		}
		b := buff[:n]
		n, err = dst.Write(b)
		if err != nil {
			t.Errorf(id, "Write failed '%s'\n", err)
			return
		}
	}
}

func (t *Tunnel) closeConn(id int) error {
	t.channels[id] <- void{}
	return nil
}

func (t *Tunnel) Close() error {
	t.lEndpoint.Close()
	t.rEndpoint.Close()
	return nil
}
