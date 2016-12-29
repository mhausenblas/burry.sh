package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"net/http"
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
}

func about() {
	fmt.Printf("This is burry in version %s\n", VERSION)
}

func get(url string, payload interface{}) error {
	c := &http.Client{Timeout: 2 * time.Second}
	r, err := c.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(payload)
}

func datadirs() {
	stateurl := strings.Join([]string{"http://", endpoint, INFRA_SVC_EXHIBITOR}, "")
	econfig := &ExhibitorState{}
	if err := get(stateurl, econfig); err != nil {
		fmt.Printf("Can't parse response from endpoint due to %s\n", err)
	} else {
		fmt.Printf("Config %#v\n", econfig)
	}
}

func walkZK() {
	zks := []string{endpoint}
	conn, _, _ := zk.Connect(zks, time.Second)
	visit(*conn, "/", rznode)
}

// rznode reaps a ZooKeeper node
func rznode(path string, val string) {
	fmt.Println(fmt.Sprintf("%s: %+v", path, val))
}

// visit visits a path in the ZooKeeper tree
func visit(conn zk.Conn, path string, fn reap) {
	fmt.Println("PROCESSING ", path)
	if children, _, err := conn.Children(path); err != nil {
		fmt.Println(fmt.Sprintf("%s", err))
		return
	} else {
		fmt.Println(path, " HAS ", len(children), " CHILDREN")
		if len(children) > 0 { // there are children
			for _, c := range children {
				fmt.Println("RAW CHILD ", c)
				newpath := ""
				if path == "/" {
					newpath = strings.Join([]string{path, c}, "")
				} else {
					newpath = strings.Join([]string{path, c}, "/")
				}
				fmt.Println("VISITING CHILD ", newpath)
				visit(conn, newpath, fn)
			}
		} else {
			if val, _, err := conn.Get(path); err != nil {
				fmt.Println(fmt.Sprintf("Can't process %s due to %s", path, err))
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
