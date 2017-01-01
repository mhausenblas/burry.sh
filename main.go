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
	VERSION        string = "0.2.0"
	BURRYFEST_FILE string = ".burryfest"
	BURRYMETA_FILE string = ".burrymeta"
)

var (
	version   bool
	overwrite bool
	// the operation burry should to carry out:
	bop  string
	BOPS = [...]string{"backup", "restore"}
	// the type of infra service to back up or restore:
	isvc           string
	INFRA_SERVICES = [...]string{"etcd", "zk"}
	// the infra service endpoint to use:
	endpoint string
	// the storage target to use:
	starget         string
	STORAGE_TARGETS = [...]string{"tty", "local", "s3", "minio"}
	// the backup and restore manifest to use:
	brf      Burryfest
	ErrNoBFF = errors.New("no manifest found")
	cred     string
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
	flag.StringVar(&bop, "operation", BOPS[0], fmt.Sprintf("The operation to carry out. Supported values are %v", BOPS))
	flag.StringVar(&isvc, "isvc", "zk", fmt.Sprintf("The type of infra service to back up or restore. Supported values are %v", INFRA_SERVICES))
	flag.StringVar(&endpoint, "endpoint", "", fmt.Sprintf("The infra service HTTP API endpoint to use. Example: localhost:8181 for Exhibitor"))
	flag.StringVar(&starget, "target", "tty", fmt.Sprintf("The storage target to use. Supported values are %v", sst))
	flag.StringVar(&cred, "credentials", "", fmt.Sprintf("The credentials to use. Example: s3.amazonaws.com,ACCESSKEYID=...,SECRETACCESSKEY=..."))

	flag.Usage = func() {
		fmt.Printf("Usage: burry %s|%s [args]\n\n", BOPS[0], BOPS[1])
		fmt.Println("Arguments:")
		flag.PrintDefaults()
	}
	flag.Parse()
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("Selected operation: %s", bop))

	if envd := os.Getenv("DEBUG"); envd != "" {
		log.SetLevel(log.DebugLevel)
	}
	c := parsecred()
	if overwrite {
		brf = Burryfest{InfraService: isvc, Endpoint: endpoint, StorageTarget: starget, Creds: c}
	} else {
		err := errors.New("")
		bfpath := ""
		if err, bfpath, brf = loadbf(); err != nil {
			if err == ErrNoBFF {
				brf = Burryfest{InfraService: isvc, Endpoint: endpoint, StorageTarget: starget, Creds: c}
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
	switch bop {
	case BOPS[0]: // backup
		switch brf.InfraService {
		case "zk":
			success = walkZK()
		case "etcd":
			success = walkETCD()
		default:
			log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	case BOPS[1]: // restore
		switch brf.InfraService {
		case "zk":
			success = true
		case "etcd":
			success = true
			log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Restoring etcd is not yet implemented"))
		default:
			log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	default:
		log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("%s is not a valid operation", bop))
		flag.Usage()
		os.Exit(2)
	}

	if success {
		if err := writebf(); err != nil {
			log.WithFields(log.Fields{"func": "main"}).Fatal(fmt.Sprintf("Something went wrong when I tried to create the burry manifest file: %s ", err))
		}
		log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Operation successfully completed."))
	} else {
		log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Operation completed with error(s)."))
		flag.Usage()
		os.Exit(1)
	}
}
