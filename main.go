package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	flag "github.com/ogier/pflag"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION        string = "0.2.0"
	BURRYFEST_FILE string = ".burryfest"
	BURRYMETA_FILE string = ".burrymeta"
	CONTENT_FILE   string = "content"
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
	zkconn   *zk.Conn
	// the storage target to use:
	starget         string
	STORAGE_TARGETS = [...]string{"tty", "local", "s3", "minio"}
	// the backup and restore manifest to use:
	brf      Burryfest
	ErrNoBFF = errors.New("no manifest found")
	cred     string
	// local scratch base directory
	based string
	// the snapshot ID
	snapshotid string
)

// reap function types take a node path
// and a value as parameters and perform
// some side effect, such as storing it.
// see for example aux.go#reapsimple()
type reap func(string, string)

func init() {
	sst := STORAGE_TARGETS[:]
	sort.Strings(sst)
	flag.BoolVarP(&version, "version", "v", false, "Display version information and exit.")
	flag.BoolVarP(&overwrite, "overwrite", "w", false, "Make command line values overwrite manifest values.")
	flag.StringVarP(&bop, "operation", "o", BOPS[0], fmt.Sprintf("The operation to carry out.\n\tSupported values are %v", BOPS))
	flag.StringVarP(&isvc, "isvc", "i", "zk", fmt.Sprintf("The type of infra service to back up or restore.\n\tSupported values are %v", INFRA_SERVICES))
	flag.StringVarP(&endpoint, "endpoint", "e", "", fmt.Sprintf("The infra service HTTP API endpoint to use.\n\tExample: localhost:8181 for Exhibitor"))
	flag.StringVarP(&starget, "target", "t", "tty", fmt.Sprintf("The storage target to use.\n\tSupported values are %v", sst))
	flag.StringVarP(&cred, "credentials", "c", "", fmt.Sprintf("The credentials to use in format STORAGE_TARGET_ENDPOINT,KEY1=VAL1,...KEYn=VALn.\n\tExample: s3.amazonaws.com,AWS_ACCESS_KEY_ID=...,AWS_SECRET_ACCESS_KEY=..."))
	flag.StringVarP(&snapshotid, "snapshot", "s", "", fmt.Sprintf("The ID of the snapshot.\n\tExample: 1483193387"))

	flag.Usage = func() {
		fmt.Printf("Usage: burry [args]\n\n")
		fmt.Println("Arguments:")
		flag.PrintDefaults()
	}
	flag.Parse()

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
	if snapshotid == "" {
		snapshotid = based
	} else {
		based = snapshotid
	}
}

func main() {
	success := false
	if version {
		about()
		os.Exit(0)
	}
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("Selected operation: %s", strings.ToUpper(bop)))
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("My config: %+v", brf))

	switch bop {
	case BOPS[0]: // backup
		switch brf.InfraService {
		case "zk":
			success = backupZK()
		case "etcd":
			success = backupETCD()
		default:
			log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	case BOPS[1]: // restore
		switch brf.InfraService {
		case "zk":
			success = restoreZK()
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
		log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Operation successfully completed. The snapshot ID is: %s", snapshotid))
	} else {
		log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Operation completed with error(s)."))
		flag.Usage()
		os.Exit(1)
	}
}
