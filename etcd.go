package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"time"
)

// backupETCD walks an etcd tree, applying
// a reap function per leaf node visited
func backupETCD() bool {
	if brf.Endpoint == "" {
		return false
	}
	cfg := client.Config{
		Endpoints:               []string{"http://" + brf.Endpoint},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, _ := client.New(cfg)
	kapi := client.NewKeysAPI(c)
	// use the etcd API to visit each node and store
	// the values in the local filesystem:
	visitETCD(kapi, "/", reapsimple)
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// create an archive file of the node's values:
		res := arch()
		// transfer to remote, if applicable:
		toremote(res)
	}
	return true
}

// visitETCD visits a path in the etcd tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visitETCD(kapi client.KeysAPI, path string, fn reap) {
	log.WithFields(log.Fields{"func": "visitETCD"}).Debug(fmt.Sprintf("On node %s", path))
	copts := client.GetOptions{
		Recursive: true,
		Sort:      false,
		Quorum:    true,
	}
	if resp, err := kapi.Get(context.Background(), path, &copts); err != nil {
		log.WithFields(log.Fields{"func": "visitETCD"}).Error(fmt.Sprintf("%s", err))
	} else {
		if resp.Node.Dir { // there are children
			log.WithFields(log.Fields{"func": "visitETCD"}).Debug(fmt.Sprintf("%s has %d children", path, len(resp.Node.Nodes)))
			for _, node := range resp.Node.Nodes {
				log.WithFields(log.Fields{"func": "visitETCD"}).Debug(fmt.Sprintf("Next visiting child %s", node.Key))
				visitETCD(kapi, node.Key, fn)
			}
		} else { // we're on a leaf node
			fn(resp.Node.Key, string(resp.Node.Value))
		}
	}
}
