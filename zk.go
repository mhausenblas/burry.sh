package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

// backupZK walks a ZooKeeper tree, applying
// a reap function per leaf node visited
func backupZK() bool {
	if brf.Endpoint == "" {
		return false
	}
	zks := []string{brf.Endpoint}
	conn, _, _ := zk.Connect(zks, time.Second)
	// use the ZK API to visit each node and store
	// the values in the local filesystem:
	visitZK(*conn, "/", reapsimple)
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// create an archive file of the node's values:
		res := arch()
		// transfer to remote, if applicable:
		toremote(res)
	}
	return true
}

// visitZK visits a path in the ZooKeeper tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visitZK(conn zk.Conn, path string, fn reap) {
	log.WithFields(log.Fields{"func": "visitZK"}).Debug(fmt.Sprintf("On node %s", path))
	if children, _, err := conn.Children(path); err != nil {
		log.WithFields(log.Fields{"func": "visitZK"}).Error(fmt.Sprintf("%s", err))
		return
	} else {
		log.WithFields(log.Fields{"func": "visitZK"}).Debug(fmt.Sprintf("%s has %d children", path, len(children)))
		if len(children) > 0 { // there are children
			for _, c := range children {
				newpath := ""
				if path == "/" {
					newpath = strings.Join([]string{path, c}, "")
				} else {
					newpath = strings.Join([]string{path, c}, "/")
				}
				log.WithFields(log.Fields{"func": "visitZK"}).Debug(fmt.Sprintf("Next visiting child %s", newpath))
				visitZK(conn, newpath, fn)
			}
		} else { // we're on a leaf node
			if val, _, err := conn.Get(path); err != nil {
				log.WithFields(log.Fields{"func": "visitZK"}).Error(fmt.Sprintf("%s", err))
			} else {
				fn(path, string(val))
			}
		}
	}
}

func restoreZK() bool {
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// transfer from remote, if applicable:

		// unarchive:

		// traverse directory and insert as per strategy:
		//  diff: only non-existing nodes will be inserted
		//  complete: overwrite existing nodes
		return true
	} else { // can't restore from TTY
		return false
	}
}
