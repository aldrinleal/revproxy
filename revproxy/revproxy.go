package main

import (
	"bitbucket.org/aldrinleal/revproxy"
	"flag"
)

var (
	port        int
	discovery   string
	defaultHost string
)

func main() {
	flag.IntVar(&port, "port", 8080, "Port to Listen")
	flag.StringVar(&discovery, "discovery", "", "etcd discovery url")
	flag.StringVar(&defaultHost, "defaultHost", "127.0.0.1", "Default Host to Forward")

	flag.Parse()

	revproxy.Port(port)
	revproxy.LoadEndpoints(flag.Args())

	revproxy.Discovery(discovery)

	revproxy.Run()
}
