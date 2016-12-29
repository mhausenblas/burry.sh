package main

import (
	"flag"
	"fmt"
	// "github.com/samuel/go-zookeeper/zk"
	"encoding/json"
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
	ExhibitorConfig
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
	get(stateurl, *econfig)
	fmt.Println("Config %+v ", econfig)
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
	datadirs()
}
