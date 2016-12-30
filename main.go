package main

import (
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION             string = "0.1.0"
	INFRA_SVC_EXHIBITOR string = "/exhibitor/v1/config/get-state"
	BURRYFEST_FILE      string = ".burryfest"
)

var (
	version   bool
	overwrite bool
	// the type of infra service to back up or restore:
	isvc           string
	INFRA_SERVICES = [...]string{"etcd", "zk"}
	// the infra service endpoint to use:
	endpoint string
	// the storage target to use:
	starget         string
	STORAGE_TARGETS = [...]string{"tty", "local"}
	// the backup and restore manifest to use:
	brf      Burryfest
	ErrNoBFF = errors.New("no manifest found")
	// local scratch base directory
	based string
)

// reap function types take a path and
// a value as parameters
type reap func(string, string)

func init() {
	sst := STORAGE_TARGETS[:]
	sort.Strings(sst)
	flag.BoolVar(&version, "version", false, "Display version information")
	flag.BoolVar(&overwrite, "overwrite", false, "Command line values overwrite manifest values")
	flag.StringVar(&isvc, "isvc", "zk", fmt.Sprintf("The type of infra service to back up or restore. Supported values are %v", INFRA_SERVICES))
	flag.StringVar(&endpoint, "endpoint", "", fmt.Sprintf("The infra service HTTP API endpoint to use. Example: localhost:8181 for Exhibitor"))
	flag.StringVar(&starget, "target", "tty", fmt.Sprintf("The storage target to use. Supported values are %v", sst))

	flag.Usage = func() {
		fmt.Printf("Usage: %s [args]\n\n", os.Args[0])
		fmt.Println("Arguments:")
		flag.PrintDefaults()
	}
	flag.Parse()
	if envd := os.Getenv("DEBUG"); envd != "" {
		log.SetLevel(log.DebugLevel)
	}
	if overwrite {
		brf = Burryfest{InfraService: isvc, Endpoint: endpoint, StorageTarget: starget, Credentials: ""}
	} else {
		err := errors.New("")
		if err, brf = loadbf(); err != nil {
			if err == ErrNoBFF {
				brf = Burryfest{InfraService: isvc, Endpoint: endpoint, StorageTarget: starget, Credentials: ""}
			}
		}
	}
	based = strconv.FormatInt(time.Now().Unix(), 10)
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("My config: %+v", brf))
}

// walkZK walks a ZooKeeper tree
func walkZK() {
	zks := []string{brf.Endpoint}
	conn, _, _ := zk.Connect(zks, time.Second)
	if err := writebf(); err != nil {
		log.WithFields(log.Fields{"func": "walkZK"}).Fatal(fmt.Sprintf("Something went wrong when I tried to create the burry manifest file: %s ", err))
	} else {
		visit(*conn, "/", rznode)
		arch()
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

// visit visits a path in the ZooKeeper tree
// and applies the reap function fn on the node
// at the path if it is a leaf node
func visit(conn zk.Conn, path string, fn reap) {
	log.WithFields(log.Fields{"func": "visit"}).Debug(fmt.Sprintf("On node %s", path))
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
				visit(conn, newpath, fn)
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

func main() {
	if version {
		about()
		os.Exit(0)
	}
	walkZK()
}
