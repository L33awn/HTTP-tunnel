package main

import (
	"flag"
)

func main() {

	pLocalAddr := flag.String("l", ":9999", "local address")
	pRemoteAddr := flag.String("r", "localhost:80", "remote address")
	flag.Parse()
}
