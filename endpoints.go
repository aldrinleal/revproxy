package revproxy

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
)

var (
	stopChannel chan bool
	errCh       chan error
	receiver    chan *etcd.Response
)

var ETCD_PREFIX = "/apps/revproxy/apps"

const (
	put = iota
	del = iota
)

func updateNode(op int, node etcd.Node) {
	k := node.Key[1+len(ETCD_PREFIX):]
	var message string
	switch op {
	case put:
		message = fmt.Sprintf("/%s:%s", k, node.Value)
	case del:
		message = fmt.Sprintf("/%s:0", k)
	}
	endpointChannel <- message
}

func EtcdPrefix(newPrefix string) {
	ETCD_PREFIX = newPrefix
}

func StartEtcd() {
	if nil == client {
		return
	}

	receiver = make(chan *etcd.Response)
	stopChannel := make(chan bool)
	errCh := make(chan error, 1)

	response, error := client.Get(ETCD_PREFIX, false, true)

	if nil != error {
		panic(error)
	}

	go func() {
		_, err := client.Watch(ETCD_PREFIX, 0, true, receiver, stopChannel)
		errCh <- err
	}()

	go func() {
		for {
			select {
			case update := <-receiver:
				op := put

				if "delete" == update.Action {
					op = del
				}

				fmt.Println("Received request for op:", op, "; node:", *update.Node)

				updateNode(op, *update.Node)
			case err := <-errCh:
				fmt.Println("Error:", err)
			}
		}
		fmt.Println("Finishing")
	}()

	for _, n := range response.Node.Nodes {
		if n.Dir {
			continue
		}

		updateNode(put, *n)
	}
}

func Stop() {
	if nil == client {
		return
	}

	stopChannel <- true

	close(errCh)
	close(stopChannel)
	close(receiver)
}
