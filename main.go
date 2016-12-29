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
)

type ExhibitorState struct {
	ExhibitorConfig ExhibitorConfig `json:"config"`
}

type ExhibitorConfig struct {
	LogDir       string `json:"logIndexDirectory"`
	SnapshotsDir string `json:"zookeeperDataDirectory"`
}

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
	if children, stat, err := conn.Children("/"); err != nil {
		fmt.Println(fmt.Sprintf("Can't find znodes due to %s", err))
	} else {
		fmt.Println(fmt.Sprintf("%+v - %+v", children, stat))
		// for _, c := range children {
		// 	if _, _, err := conn.Get(c); err != nil {
		// 	}
		// }
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
