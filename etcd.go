package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// backupETCD walks an etcd tree, applying
// a reap function per leaf node visited
func backupETCD() bool {
	if brf.Endpoint == "" {
		return false
	}
	cfg := etcd.Config{
		Endpoints:               []string{"http://" + brf.Endpoint},
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, _ := etcd.New(cfg)
	kapi = etcd.NewKeysAPI(c)
	// use the etcd API to visit each node and store
	// the values in the local filesystem:
	visitETCD("/", reapsimple)
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
func visitETCD(path string, fn reap) {
	log.WithFields(log.Fields{"func": "visitETCD"}).Debug(fmt.Sprintf("On node %s", path))
	copts := etcd.GetOptions{
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
				visitETCD(node.Key, fn)
			}
		} else { // we're on a leaf node
			fn(resp.Node.Key, string(resp.Node.Value))
		}
	}
}

func restoreETCD() bool {
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// transfer from remote, if applicable:
		a := fromremote()
		// unarchive:
		s := unarch(a)
		defer func() {
			_ = os.RemoveAll(s)
		}()
		cfg := etcd.Config{
			Endpoints:               []string{"http://" + brf.Endpoint},
			Transport:               etcd.DefaultTransport,
			HeaderTimeoutPerRequest: time.Second,
		}
		c, _ := etcd.New(cfg)
		kapi = etcd.NewKeysAPI(c)
		// walk the snapshot directory and use the etcd API to
		// restore keys from the local filesystem - note that
		// only non-existing keys will be created:
		if err := filepath.Walk(s, visitETCDReverse); err != nil {
			log.WithFields(log.Fields{"func": "restoreETCD"}).Error(fmt.Sprintf("%s", err))
			return false
		}
	} else { // can't restore from TTY
		return false
	}
	return true
}

func visitETCDReverse(path string, f os.FileInfo, err error) error {
	if f.Name() == BURRYMETA_FILE || f.Name() == snapshotid {
		return nil
	} else {
		cwd, _ := os.Getwd()
		base, _ := filepath.Abs(filepath.Join(cwd, snapshotid))
		key, _ := filepath.Rel(base, path)
		// append the root "/" to make it a key and unescape ":"
		key = "/" + strings.Replace(key, "BURRY_ESC_COLON", ":", -1)
		if f.IsDir() {
			cfile, _ := filepath.Abs(filepath.Join(path, CONTENT_FILE))
			if _, eerr := os.Stat(cfile); eerr == nil { // there is a content file at this path
				log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Attempting to insert %s as leaf key", key))
				if c, cerr := readc(cfile); cerr != nil {
					log.WithFields(log.Fields{"func": "visitETCDReverse"}).Error(fmt.Sprintf("%s", cerr))
					return cerr
				} else {
					if _, kerr := kapi.Set(context.Background(), key, string(c), &etcd.SetOptions{Dir: false, PrevExist: etcd.PrevNoExist}); kerr == nil {
						log.WithFields(log.Fields{"func": "visitETCDReverse"}).Info(fmt.Sprintf("Restored %s", key))
						log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Value: %s", c))
						numrestored = numrestored + 1
					}
				}
			} else {
				log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Attempting to insert %s as a non-leaf key", key))
				if _, kerr := kapi.Set(context.Background(), key, "", &etcd.SetOptions{Dir: true, PrevExist: etcd.PrevNoExist}); kerr == nil {
					log.WithFields(log.Fields{"func": "visitETCDReverse"}).Info(fmt.Sprintf("Restored %s", key))
					numrestored = numrestored + 1
				}
			}
		}
		log.WithFields(log.Fields{"func": "visitETCDReverse"}).Debug(fmt.Sprintf("Visited %s", key))
	}
	return nil
}
