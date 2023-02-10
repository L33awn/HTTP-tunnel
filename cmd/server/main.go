package main

import (
	"flag"
	"fmt"
	"os"
	"tunnel"
)

func main() {

	pLocalAddr := flag.String("l", "", "local address")
	pRemoteAddr := flag.String("r", "", "remote address")

	flag.Parse()

	if *pLocalAddr == "" || *pRemoteAddr == "" {
		fmt.Printf("Usage: %s -l proto://address -r proto://address\n", os.Args[0])
		os.Exit(-1)
	}

	t := tunnel.NewTunnel()
}
