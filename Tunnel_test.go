package tunnel

import (
	"log"
	"net/url"
	"testing"
)

func TestTunnel(t *testing.T) {
	lu1, _ := url.Parse("tcp://127.0.0.1:6666")
	ru1, _ := url.Parse("http://127.0.0.1:7777")
	t1, err := NewTunnel(
		lu1, ru1,
	)
	if err != nil {
		log.Fatalln(err)
	}
	lu2, _ := url.Parse("http://127.0.0.1:7777")
	ru2, _ := url.Parse("tcp://127.0.0.1:8888")
	t2, err := NewTunnel(
		lu2, ru2,
	)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		err = t1.Start()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	err = t2.Start()
	if err != nil {
		log.Fatalln(err)
	}
}
