package revproxy

import (
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http"
	"os"
)

// GetEtcdHosts returns the lists of etcd hosts in the cluster
func GetEtcdHosts(discoveryURL string) ([]string, error) {
	// Part #1: Query the Discovery Server to get the Peer Lists
	response, err := http.Get(discoveryURL + "?recursive=true")

	if nil != err {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)

	if nil != err {
		return nil, err
	}

	json, _ := simplejson.NewJson(contents)

	nodes, err := json.GetPath("node", "nodes").Array()

	if nil != err {
		return nil, err
	}

	// Parses/Picks/builds a list of peers addresses - which we'll use later
	strNodes := make([]string, len(nodes))

	for i, v := range nodes {
		v := v.(map[string]interface{})

		strNodes[i] = v["value"].(string)
	}

	contents = nil

	/*
	 * Attempts to get the first relation of machines. It doesn't care too much if it fails (provided we have a cluster), so we get more lax with error handling
	 */
	for _, v := range strNodes {
		response, err = http.Get(v + "/v2/admin/machines")

		if nil != err {
			continue
		}

		defer response.Body.Close()

		contents, err = ioutil.ReadAll(response.Body)

		if nil != err {
			continue
		}
	}

	if nil == contents {
		return nil, err
	}

	// Now'rell properly parse and return the values of "clientURL", as-is

	json, _ = simplejson.NewJson(contents)

	nodes, err = json.Array()

	if nil != err {
		return nil, err
	}

	result := make([]string, len(nodes))

	for i, v := range nodes {
		v := v.(map[string]interface{})

		result[i] = v["clientURL"].(string)
	}

	return result, nil
}

func main() {
	result, _ := GetEtcdHosts(os.Args[1])

	fmt.Println(result)
}
