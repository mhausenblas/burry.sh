package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"time"
)

// walkETCD walks an etcd tree, applying
// a reap function per node
func walkETCD() bool {
	cfg := client.Config{
		Endpoints:               []string{"http://" + brf.Endpoint},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, _ := client.New(cfg)
	kapi := client.NewKeysAPI(c)
	if resp, err := kapi.Get(context.Background(), "/foo", nil); err != nil {
		log.WithFields(log.Fields{"func": "rekey"}).Error(fmt.Sprintf("%s", err))
	} else {
		log.WithFields(log.Fields{"func": "rekey"}).Info(fmt.Sprintf("metadata: %v, value: %q\n", resp, resp.Node.Value))
	}
	return true
}

// visitETCD visits a path in the etcd tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visitETCD(path string, fn reap) {
}

// rekey reaps an etcd key.
// note that the actual processing is determined by
// the storage target
func rekey(path string, val string) {
	switch lookupst(brf.StorageTarget) {
	case 0: // TTY
		log.WithFields(log.Fields{"func": "rekey"}).Info(fmt.Sprintf("%s:", path))
		log.WithFields(log.Fields{"func": "rekey"}).Debug(fmt.Sprintf("%v", val))
	case 1: // local storage
		store(path, val)
	default:
		log.WithFields(log.Fields{"func": "rekey"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
	}
}
