package tunnel

import (
	"fmt"
	"net/http"
)

type HttpConn struct {
	id      int
	address string
	WriteC  chan []byte
}

func NewHttpConn(address string, id int) (hc *HttpConn, err error) {
	res, err := http.Get(fmt.Sprintf("http://%s%s?id=%d", address, TEST_ENDPOINT, id))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status Code is not 200")
	}
	hc = &HttpConn{
		id:      id,
		address: address,
		WriteC:  make(chan []byte),
	}
	return
}

func (hc *HttpConn) Read(p []byte) (n int, err error) {
	res, err := http.Get(fmt.Sprintf("http://%s%s?id=%d", hc.address, READ_ENDPOINT, hc.id))
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Status Code is not 200")
	}
	return res.Body.Read(p)
}

func (hc *HttpConn) Write(p []byte) (n int, err error) {
	res, err := http.Get(fmt.Sprintf("http://%s%s?id=%d", hc.address, REQUEST_READ_ENDPOINT, hc.id))
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Status Code is not 200")
	}
	sz := len(p)
	buf := make([]byte, sz)
	copy(buf, p)
	hc.WriteC <- buf
	return sz, nil
}

func (hc *HttpConn) Close() error {

}
