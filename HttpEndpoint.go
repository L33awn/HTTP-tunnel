package tunnel

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	NEW_CONN_ENDPOINT = "/new"
	CLOSE_ENDPOINT    = "/close"
	// REQUEST_READ_ENDPOINT = "/req_read"
	READ_ENDPOINT = "/read"
	TEST_ENDPOINT = "/test"

	REQUEST_INTERVAL = 2 * time.Second
)

type HttpEndpoint struct {
	address  string
	tunnel   *Tunnel
	serveMux *http.ServeMux
	sessions map[int]*HttpConn
	server   *http.Server
	closeC   chan error
}

func newHttpEndpoint(u *url.URL, t *Tunnel, isLocal bool) (he *HttpEndpoint, err error) {
	lAddress := ":0"
	if isLocal {
		lAddress = u.Host
	}
	l, err := net.Listen("tcp", lAddress)
	if err != nil {
		return nil, err
	}
	serveMux := &http.ServeMux{}
	he = &HttpEndpoint{
		address:  u.Host,
		tunnel:   t,
		serveMux: serveMux,
		sessions: make(map[int]*HttpConn),
		server: &http.Server{
			Addr:    l.Addr().String(),
			Handler: serveMux,
		},
		closeC: make(chan error),
	}
	err = he.initHttpServer()
	if err != nil {
		return nil, err
	}
	go func() {
		he.closeC <- he.server.Serve(l)
	}()
	return
}

func (he *HttpEndpoint) ListenAndServe() (err error) {
	return <-he.closeC
}

func (he *HttpEndpoint) Dial() (io.ReadWriteCloser, error) {
	id := int(rand.Int31())
	q := &url.Values{
		"address": []string{he.server.Addr},
		"id":      []string{strconv.Itoa(id)},
	}
	u := fmt.Sprintf("http://%s%s?%s", he.address, NEW_CONN_ENDPOINT, q.Encode())
	fmt.Println("HttpEndpoint dialing", u)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status Code is not 200")
	}
	conn, err := NewHttpConn(he.address, id)
	if err != nil {
		return nil, err
	}
	he.sessions[id] = conn
	return conn, nil
}

func (he *HttpEndpoint) Close() (err error) {
	return he.server.Shutdown(context.TODO())
}

func (he *HttpEndpoint) initHttpServer() (err error) {
	he.serveMux.HandleFunc(NEW_CONN_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		// ip := r.URL.Query().Get("ip")
		// port := r.URL.Query().Get("port")
		queries := r.URL.Query()
		address := queries.Get("address")
		strId := queries.Get("id")
		if address == "" || strId == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id, err := strconv.Atoi(strId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conn, err := NewHttpConn(address, id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		he.sessions[id] = conn
		go he.tunnel.newConnection(conn)
	})
	he.serveMux.HandleFunc(READ_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		strId := r.URL.Query().Get("id")
		if strId == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id, err := strconv.Atoi(strId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println("Handling Reading from", he.sessions[id].WriteC)
		dataToWrite := <-he.sessions[id].WriteC
		if dataToWrite == nil {
			return
		}
		log.Println("Writing", dataToWrite)
		_, err = w.Write(dataToWrite)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	he.serveMux.HandleFunc(TEST_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		strId := r.URL.Query().Get("id")
		if strId == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id, err := strconv.Atoi(strId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// if he.sessions[id] == nil {

		// }
		w.Write([]byte(fmt.Sprintf("%d", id)))
	})
	he.serveMux.HandleFunc(CLOSE_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		he.sessions[id].Close()
	})
	return nil
}
