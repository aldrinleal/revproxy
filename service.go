package revproxy

import (
	"regexp"
	"fmt"
	"strconv"
	"log"
	"net/http"
	"sort"
	"net/http/httputil"
	"net/url"
)

type _Endpoint struct {
	prefix  string
	port    int
	re *regexp.Regexp
	targetURL *url.URL
	handler http.Handler
}

var endpoints []_Endpoint

var port int

func Port(portToSet int) {
	port = portToSet
}

func make_endpoint(loc string) (*_Endpoint, error) {
	PATTERN_ENDPOINT := regexp.MustCompile(`/(\w*):(\d+)`)

	elements := PATTERN_ENDPOINT.FindStringSubmatch(loc)

	if 3 != len(elements) {
		return nil, fmt.Errorf("Invalid Path '%s'. Must match /<path>:<port>", loc)
	}

	prefix := elements[1]
	port, _ := strconv.Atoi(elements[2])

	urlAsString := fmt.Sprintf("http://127.0.0.1:%d/", port)

	targetUrl, err := url.Parse(urlAsString)

	if nil != err {
		log.Fatal(err)
	}

	handler := make_handler(targetUrl)

	path_re := regexp.MustCompile(fmt.Sprintf("^/%s(/.*)?", prefix))

	new_endpoint := &_Endpoint{ "/" + prefix, port, path_re, targetUrl, handler }

	return new_endpoint, nil
}

func make_handler(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery+req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery+"&"+req.URL.RawQuery
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

type _ByLen []_Endpoint

func (a _ByLen) Len() int { return len(a) }
func (a _ByLen) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a _ByLen) Less(i, j int) bool { return len(a[i].prefix) > len(a[j].prefix) }

func LoadEndpoints(args []string) {
	endpoints = make([]_Endpoint, len(args))

	for i, v := range args {
		new_endpoint, err := make_endpoint(v)

		if nil != err {
			log.Fatal(err)
		}

		endpoints[i] = *new_endpoint
	}

	sort.Sort(_ByLen(endpoints))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	matchingServerOf := func(url *url.URL) (http.Handler, bool) {
		for _, v := range endpoints {
			if v.re.MatchString(url.Path) {
				log.Println(" -> ", v.targetURL.RequestURI())
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
