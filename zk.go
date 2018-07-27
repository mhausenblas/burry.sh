package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"path/filepath"
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
	zkconn, _, _ = zk.Connect(zks, time.Second)
	// use the ZK API to visit each node and store
	// the values in the local filesystem:
	visitZK("/", reapsimple)
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// create an archive file of the node's values:
		a := arch()
		// transfer to remote, if applicable:
		toremote(a)
	}
	return true
}

// visitZK visits a path in the ZooKeeper tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visitZK(path string, fn reap) {
	log.WithFields(log.Fields{"func": "visitZK"}).Debug(fmt.Sprintf("On node %s", path))
	if children, _, err := zkconn.Children(path); err != nil {
		log.WithFields(log.Fields{"func": "visitZK"}).Error(fmt.Sprintf("%s", err))
		return
	} else {
		log.WithFields(log.Fields{"func": "visitZK"}).Debug(fmt.Sprintf("%s has %d children", path, len(children)))

		if val, _, err := zkconn.Get(path); err != nil {
			log.WithFields(log.Fields{"func": "visitZK"}).Error(fmt.Sprintf("%s", err))
		} else {
			fn(path, string(val))
		}

		if len(children) > 0 { // there are children
			for _, c := range children {
				newpath := ""
				if path == "/" {
					newpath = strings.Join([]string{path, c}, "")
				} else {
					newpath = strings.Join([]string{path, c}, "/")
				}
				log.WithFields(log.Fields{"func": "visitZK"}).Debug(fmt.Sprintf("Next visiting child %s", newpath))
				visitZK(newpath, fn)
			}
		}
	}
}

func restoreZK() bool {
	if lookupst(brf.StorageTarget) > 0 { // non-TTY, actual storage
		// transfer from remote, if applicable:
		a := fromremote()
		// unarchive:
		s := unarch(a)
		defer func() {
			_ = os.RemoveAll(s)
		}()
		zks := []string{brf.Endpoint}
		zkconn, _, _ = zk.Connect(zks, time.Second)
		zkconn.SetLogger(log.StandardLogger())
		// walk the snapshot directory and use the ZK API to
		// restore znodes from the local filesystem - note that
		// only non-existing znodes will be created:
		if err := filepath.Walk(s, visitZKReverse); err != nil {
			log.WithFields(log.Fields{"func": "restoreZK"}).Error(fmt.Sprintf("%s", err))
			return false
		}
	} else { // can't restore from TTY
		return false
	}
	return true
}

func visitZKReverse(path string, f os.FileInfo, err error) error {
	if f.Name() == BURRYMETA_FILE || f.Name() == snapshotid {
		return nil
	} else {
		cwd, _ := os.Getwd()
		base, _ := filepath.Abs(filepath.Join(cwd, snapshotid))
		znode, _ := filepath.Rel(base, path)
		// append the root "/" to make it a znode and unescape ":"
		znode = "/" + strings.Replace(znode, "BURRY_ESC_COLON", ":", -1)
		if pathpresent, _, err := zkconn.Exists(znode); err != nil {
			log.WithFields(log.Fields{"func": "visitZKReverse"}).Error(fmt.Sprintf("%s", err))
			return err
		} else {
			if pathpresent {
				log.WithFields(log.Fields{"func": "visitZKReverse"}).Debug(fmt.Sprintf("znode %s exists already, skipping it", znode))
			} else {
				if f.IsDir() {
					cfile, _ := filepath.Abs(filepath.Join(path, CONTENT_FILE))
					if _, eerr := os.Stat(cfile); eerr == nil { // there is a content file at this path
						log.WithFields(log.Fields{"func": "visitZKReverse"}).Debug(fmt.Sprintf("Attempting to insert %s as leaf znode", znode))
						if c, cerr := readc(cfile); cerr != nil {
							log.WithFields(log.Fields{"func": "visitZKReverse"}).Error(fmt.Sprintf("%s", cerr))
							return cerr
						} else {
							if _, zerr := zkconn.Create(znode, c, 0, zk.WorldACL(zk.PermAll)); zerr != nil {
								log.WithFields(log.Fields{"func": "visitZKReverse"}).Error(fmt.Sprintf("%s", zerr))
								return zerr
							} else {
								log.WithFields(log.Fields{"func": "visitZKReverse"}).Info(fmt.Sprintf("Restored %s", znode))
								log.WithFields(log.Fields{"func": "visitZKReverse"}).Debug(fmt.Sprintf("Value: %s", c))
								numrestored = numrestored + 1
							}
						}
					} else {
						log.WithFields(log.Fields{"func": "visitZKReverse"}).Debug(fmt.Sprintf("Attempting to insert %s as a non-leaf znode", znode))
						if _, zerr := zkconn.Create(znode, []byte{}, 0, zk.WorldACL(zk.PermAll)); zerr != nil {
							log.WithFields(log.Fields{"func": "visitZKReverse"}).Error(fmt.Sprintf("%s", zerr))
							return zerr
						} else {
							log.WithFields(log.Fields{"func": "visitZKReverse"}).Info(fmt.Sprintf("Restored %s", znode))
							numrestored = numrestored + 1
						}
					}
				}
			}
		}
		log.WithFields(log.Fields{"func": "visitZKReverse"}).Debug(fmt.Sprintf("Visited %s", znode))
	}
	return nil
}
