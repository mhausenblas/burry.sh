package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

// walkZK walks a ZooKeeper tree, applying
// a reap function per node
func walkZK() bool {
	zks := []string{brf.Endpoint}
	conn, _, _ := zk.Connect(zks, time.Second)
	visitZK(*conn, "/", rznode)
	arch()
	return true
}

// visitZK visits a path in the ZooKeeper tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visitZK(conn zk.Conn, path string, fn reap) {
	log.WithFields(log.Fields{"func": "visit"}).Info(fmt.Sprintf("On node %s", path))
	if children, _, err := conn.Children(path); err != nil {
		log.WithFields(log.Fields{"func": "visit"}).Error(fmt.Sprintf("%s", err))
		return
	} else {
		log.WithFields(log.Fields{"func": "visit"}).Debug(fmt.Sprintf("%s has %d children", path, len(children)))
		if len(children) > 0 { // there are children
			for _, c := range children {
				newpath := ""
				if path == "/" {
					newpath = strings.Join([]string{path, c}, "")
				} else {
					newpath = strings.Join([]string{path, c}, "/")
				}
				log.WithFields(log.Fields{"func": "visit"}).Debug(fmt.Sprintf("Next visiting child %s", newpath))
				visitZK(conn, newpath, fn)
			}
		} else { // we're on a leaf node
			if val, _, err := conn.Get(path); err != nil {
				log.WithFields(log.Fields{"func": "visit"}).Error(fmt.Sprintf("%s", err))
			} else {
				fn(path, string(val))
			}
		}
	}
}

// rznode reaps a ZooKeeper node.
// note that the actual processing is determined by
// the storage target
func rznode(path string, val string) {
	switch lookupst(brf.StorageTarget) {
	case 0: // TTY
		log.WithFields(log.Fields{"func": "rznode"}).Info(fmt.Sprintf("%s:", path))
		log.WithFields(log.Fields{"func": "rznode"}).Debug(fmt.Sprintf("%v", val))
	case 1: // local storage
		store(path, val)
	default:
		log.WithFields(log.Fields{"func": "rznode"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
	}
}
