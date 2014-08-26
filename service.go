// Package revproxy ...
package revproxy

import (
	"fmt"
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
var endpoints []_Endpoint

// port to bind
var port int

// Port sets the port
func Port(portToSet int) {
	port = portToSet
}

// makeEndpoint creates an endpoint
func makeEndpoint(loc string) (*_Endpoint, error) {
	endpointRe := regexp.MustCompile(`/(\w*):(\d+)`)

	elements := endpointRe.FindStringSubmatch(loc)

	if 3 != len(elements) {
		return nil, fmt.Errorf("Invalid Path '%s'. Must match /<path>:<port>", loc)
	}

	prefix := elements[1]
	port, _ := strconv.Atoi(elements[2])

	urlAsString := fmt.Sprintf("http://127.0.0.1:%d/", port)

	targetURL, err := url.Parse(urlAsString)

	if nil != err {
		log.Fatal(err)
	}

	handler := makeHandler(targetURL)

	pathRe := regexp.MustCompile(fmt.Sprintf("^/%s(/.*)?", prefix))

	newEndpoint := &_Endpoint{"/" + prefix, port, pathRe, targetURL, handler}

	return newEndpoint, nil
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

// LoadEndpoints parses the cli and builds the relation of available endpoints
func LoadEndpoints(args []string) {
	endpoints = make([]_Endpoint, len(args))

	for i, v := range args {
		newEndpoint, err := makeEndpoint(v)

		if nil != err {
			log.Fatal(err)
		}

		endpoints[i] = *newEndpoint
	}

	sort.Sort(_ByLen(endpoints))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	matchingServerOf := func(url *url.URL) (http.Handler, bool) {
		for _, v := range endpoints {
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
