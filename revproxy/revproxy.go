package main

import (
	"flag"
	"github.com/aldrinleal/revproxy"
	"os"
)

var (
	port             int
	etcdDiscoveryURL string
	defaultHost      string
	etcdPrefix       string
)

func defaultEnv(envName, defaultValue string) string {
	if "" != os.Getenv(envName) {
		return os.Getenv(envName)
	}

	return defaultValue
}

func main() {
	flag.IntVar(&port, "port", 8080, "Port to Listen")
	flag.StringVar(&etcdDiscoveryURL, "etcdDiscoveryURL", defaultEnv("ETCD_DISCOVERY_URL", ""), "etcd discovery URL")
	flag.StringVar(&defaultHost, "defaultHost", defaultEnv("REVPROXY_HOST", "127.0.0.1"), "Default Host to Forward")
	flag.StringVar(&etcdPrefix, "etcdPrefix", defaultEnv("ETCD_PREFIX", revproxy.ETCD_PREFIX), "Default ETCD Prefix")

	flag.Parse()

	revproxy.DefaultHost(defaultHost)
	revproxy.EtcdPrefix(etcdPrefix)
	revproxy.Port(port)
	if "" != etcdDiscoveryURL {
		revproxy.Discovery(etcdDiscoveryURL)
	} else {
		revproxy.LoadEndpoints(flag.Args())
	}

	revproxy.Run()
}
