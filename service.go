// Package revproxy ...
package revproxy

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sort"
	"strconv"
)

// _Endpoint
type _Endpoint struct {
	prefix    string
	port      int
	re        *regexp.Regexp
	targetURL *url.URL
	handler   http.Handler
}

// endpoints
var (
	endpoints       = make(map[string]_Endpoint)
	endpointList    []_Endpoint
	port            int
	client          *etcd.Client
	endpointChannel = make(chan string)
	defaultHost     = "127.0.0.1"
)

func DefaultHost(defaultHostToSet string) {
	defaultHost = defaultHostToSet
}

// Port sets the port
func Port(portToSet int) {
	port = portToSet
}

func init() {
	go func() {
		for {
			select {
			case endpointUpdate := <-endpointChannel:
				{
					_, err := updateEndpoint(endpointUpdate)

					if nil != err {
						fmt.Println("Error for endpoint", endpointUpdate, ":", err)
					}
				}
			}
		}
	}()
}

// updateEndpoint updates endpoints. If port is set to Zero, its removed
func updateEndpoint(loc string) (*_Endpoint, error) {
	endpointRe := regexp.MustCompile(`/(\w*):(\d+)`)

	elements := endpointRe.FindStringSubmatch(loc)

	if 3 != len(elements) {
		return nil, fmt.Errorf("Invalid Path '%s'. Must match /<path>:<port>", loc)
	}

	prefix := elements[1]

	if "default" == prefix {
		prefix = ""
	}

	port, _ := strconv.Atoi(elements[2])

	if 0 == port {
		delete(endpoints, prefix)

		fmt.Println(fmt.Sprintf("Excluding endpoint for '%s'", prefix))

		updateEndpoints()

		return nil, nil
	}

	urlAsString := fmt.Sprintf("http://%s:%d/", defaultHost, port)

	targetURL, err := url.Parse(urlAsString)

	if nil != err {
		log.Fatal(err)
	}

	handler := makeHandler(targetURL)

	pathRe := regexp.MustCompile(fmt.Sprintf("^/%s(/.*)?", prefix))

	newEndpoint := &_Endpoint{"/" + prefix, port, pathRe, targetURL, handler}

	endpoints[prefix] = *newEndpoint

	fmt.Println("Adding new endpoint: ", endpoints[prefix])

	updateEndpoints()

	return newEndpoint, nil
}

func updateEndpoints() {
	newEndpointList := make([]_Endpoint, len(endpoints))

	j := 0
	for _, v := range endpoints {
		newEndpointList[j] = v
		j++
	}

	sort.Sort(_ByLen(newEndpointList))

	endpointList = newEndpointList
}

func makeHandler(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

type _ByLen []_Endpoint

func (a _ByLen) Len() int           { return len(a) }
func (a _ByLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a _ByLen) Less(i, j int) bool { return len(a[i].prefix) > len(a[j].prefix) }

func Discovery(discoveryURL string) {
	if "" == discoveryURL {
		return
	}

	fmt.Println("Using discovery url: ", discoveryURL)

	machines, err := GetEtcdHosts(discoveryURL)

	if nil != err {
		panic(err)
	}

	fmt.Println("Using etcd endpoints: ", machines)

	client = etcd.NewClient(machines)
}

// LoadEndpoints parses the cli and builds the relation of available endpoints
func LoadEndpoints(args []string) {
	for _, v := range args {
		updateEndpoint(v)
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	matchingServerOf := func(url *url.URL) (http.Handler, bool) {
		for _, v := range endpointList {
			if v.re.MatchString(url.Path) {
				return v.handler, true
			}
		}

		return nil, false
	}

	server, found := matchingServerOf(r.URL)

	if found {
		server.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

// Run launches the HTTP Daemon
func Run() {
	StartEtcd()

	log.Println("Defined Endpoints:")

	for i, v := range endpoints {
		log.Println(i, v)
	}

	http.HandleFunc("/", proxy)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if nil != err {
		log.Fatal(err)
	}
}
