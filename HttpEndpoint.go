package tunnel

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	NEW_CONN_ENDPOINT     = "/new"
	CLOSE_ENDPOINT        = "/close"
	REQUEST_READ_ENDPOINT = "/req_read"
	READ_ENDPOINT         = "/read"
	TEST_ENDPOINT         = "/test"

	REQUEST_INTERVAL = 2 * time.Second
)

type HttpEndpoint struct {
	address  string
	tunnel   *Tunnel
	serveMux *http.ServeMux
	sessions map[int]*HttpConn
	server   *http.Server
}

func newHttpEndpoint(u *url.URL, t *Tunnel) (he *HttpEndpoint, err error) {
	address := fmt.Sprintf("%s:%s", u.Host, u.Port())
	serveMux := &http.ServeMux{}
	return &HttpEndpoint{
		address:  address,
		tunnel:   t,
		serveMux: serveMux,
		server: &http.Server{
			Addr:    address,
			Handler: serveMux,
		},
	}, nil
}

func (he *HttpEndpoint) ListenAndServe() (err error) {
	he.serveMux.HandleFunc(NEW_CONN_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		port := r.URL.Query().Get("port")
		strId := r.URL.Query().Get("id")
		if ip == "" || port == "" || strId == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id, err := strconv.Atoi(strId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conn, err := NewHttpConn(fmt.Sprintf("%s:%s", ip, port), id)
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
		_, err = w.Write(<-he.sessions[id].WriteC)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	// he.serveMux.HandleFunc(REQUEST_READ_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
	// 	strId := r.URL.Query().Get("id")
	// 	if strId == "" {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}
	// 	id, err := strconv.Atoi(strId)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}
	// 	he.sessions[id].Read()
	// })
	he.serveMux.HandleFunc(CLOSE_ENDPOINT, func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		he.sessions[id].Close()
	})
	return he.server.ListenAndServe()
}

func (he *HttpEndpoint) Dial() (io.ReadWriteCloser, error) {
	res, err := http.Get(fmt.Sprintf("http://%s%s", he.address, NEW_CONN_ENDPOINT))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status Code is not 200")
	}
	id := int(rand.Int31())
	return NewHttpConn(he.address, id)
}

func (he *HttpEndpoint) Close() (err error) {
	return he.server.Shutdown(context.TODO())
}
