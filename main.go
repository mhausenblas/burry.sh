package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"strings"
	"time"
)

const (
	VERSION             string = "0.1.0"
	INFRA_SVC_EXHIBITOR string = "/exhibitor/v1/config/get-state"
)

var (
	version bool
	// the endpoint to use
	endpoint string
	// the backup and restore manifest to use
	brm BRManifest
)

// reap function types take a path and
// a value as parameters
type reap func(string, string)

func init() {
	flag.BoolVar(&version, "version", false, "Display version information")
	flag.StringVar(&endpoint, "endpoint", "", fmt.Sprintf("The endpoint to use. This depends on the infra service you want to back up. Example: localhost:8181 for Exhibitor"))

	flag.Usage = func() {
		fmt.Printf("Usage: %s [args]\n\n", os.Args[0])
		fmt.Println("Arguments:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if envd := os.Getenv("DEBUG"); envd != "" {
		log.SetLevel(log.DebugLevel)
	}
}

func walkZK() {
	zks := []string{endpoint}
	conn, _, _ := zk.Connect(zks, time.Second)
	visit(*conn, "/", rznode)
}

// rznode reaps a ZooKeeper node
func rznode(path string, val string) {
	log.WithFields(log.Fields{"func": "rznode"}).Info(fmt.Sprintf("%s:", path))
	log.WithFields(log.Fields{"func": "rznode"}).Debug(fmt.Sprintf("%v", val))
}

// visit visits a path in the ZooKeeper tree
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
		} else { // we're on a leave node
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
	if endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	walkZK()
}
