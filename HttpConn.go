package tunnel

import (
	"fmt"
	"net/http"
)

type HttpConn struct {
	id      int
	address string
	WriteC  chan []byte
	closed  bool
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
	fmt.Println("Reading from", fmt.Sprintf("http://%s%s?id=%d", hc.address, READ_ENDPOINT, hc.id))
	res, err := http.Get(fmt.Sprintf("http://%s%s?id=%d", hc.address, READ_ENDPOINT, hc.id))
	fmt.Println("Read done")
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Status Code is not 200")
	}
	return res.Body.Read(p)
}

func (hc *HttpConn) Write(p []byte) (n int, err error) {
	// res, err := http.Get(fmt.Sprintf("http://%s%s?id=%d", hc.address, REQUEST_READ_ENDPOINT, hc.id))
	// if err != nil {
	// 	return
	// }
	// if res.StatusCode != http.StatusOK {
	// 	return 0, fmt.Errorf("Status Code is not 200")
	// }
	sz := len(p)
	if sz == 0 {
		return 0, nil
	}
	fmt.Println("writing size", sz)
	buf := make([]byte, sz)
	copy(buf, p)
	hc.WriteC <- buf
	return sz, nil
}

func (hc *HttpConn) Close() (err error) {
	if !hc.closed {
		hc.closed = true
		_, err = http.Get(fmt.Sprintf("http://%s%s?id=%d", hc.address, CLOSE_ENDPOINT, hc.id))
		close(hc.WriteC)
	}
	return err
}
