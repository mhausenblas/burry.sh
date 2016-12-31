package main

import (
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	VERSION        string = "0.1.0"
	BURRYFEST_FILE string = ".burryfest"
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
	STORAGE_TARGETS = [...]string{"tty", "local", "s3"}
	// the backup and restore manifest to use:
	brf      Burryfest
	ErrNoBFF = errors.New("no manifest found")
	// local scratch base directory
	based string
)

// reap function types take a node path
// and a value as parameters and perform
// some side effect, such as storing it.
// see for example aux.go#reapsimple()
type reap func(string, string)

func init() {
	sst := STORAGE_TARGETS[:]
	sort.Strings(sst)
	flag.BoolVar(&version, "version", false, "Display version information")
	flag.BoolVar(&overwrite, "overwrite", false, "Make command line values overwrite manifest values")
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
		bfpath := ""
		if err, bfpath, brf = loadbf(); err != nil {
			if err == ErrNoBFF {
				brf = Burryfest{InfraService: isvc, Endpoint: endpoint, StorageTarget: starget, Credentials: ""}
			} else {
				log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("Using existing burry manifest file %s", bfpath))
			}
		} else {
			log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("Using existing burry manifest file %s", bfpath))
		}
	}
	based = strconv.FormatInt(time.Now().Unix(), 10)
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("My config: %+v", brf))
}

func main() {
	success := false
	if version {
		about()
		os.Exit(0)
	}
	if err := writebf(); err != nil {
		log.WithFields(log.Fields{"func": "main"}).Fatal(fmt.Sprintf("Something went wrong when I tried to create the burry manifest file: %s ", err))
	}
	switch brf.InfraService {
	case "zk":
		success = walkZK()
	case "etcd":
		success = walkETCD()
	default:
		log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
	}
	if success {
		log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Operation successfully completed."))
	} else {
		log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Operation completed with error(s)."))
		flag.Usage()
		os.Exit(1)
	}
}
