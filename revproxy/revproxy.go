package main

import (
	"flag"
	"bitbucket.org/aldrinleal/revproxy"
)

var (
	port int
)

func main() {
	flag.IntVar(&port, "port", 8080, "Port to Listen")

	flag.Parse()

	revproxy.Port(port)
	revproxy.LoadEndpoints(flag.Args())

	revproxy.Run()
}
