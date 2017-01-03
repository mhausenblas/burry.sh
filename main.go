package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	flag "github.com/ogier/pflag"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION                 string = "0.3.0"
	BURRYFEST_FILE          string = ".burryfest"
	BURRYMETA_FILE          string = ".burrymeta"
	CONTENT_FILE            string = "content"
	BURRY_OPERATION_BACKUP  string = "backup"
	BURRY_OPERATION_RESTORE string = "restore"
	INFRA_SERVICE_ETCD      string = "etcd"
	INFRA_SERVICE_ZK        string = "zk"
	STORAGE_TARGET_TTY      string = "tty"
	STORAGE_TARGET_LOCAL    string = "local"
	STORAGE_TARGET_S3       string = "s3"
	STORAGE_TARGET_MINIO    string = "minio"
	REMOTE_ARCH_FILE        string = "latest.zip"
	REMOTE_ARCH_TYPE        string = "application/zip"
)

var (
	version   bool
	overwrite bool
	// the operation burry should to carry out:
	bop  string
	BOPS = [...]string{BURRY_OPERATION_BACKUP, BURRY_OPERATION_RESTORE}
	// the type of infra service to back up or restore:
	isvc           string
	INFRA_SERVICES = [...]string{INFRA_SERVICE_ETCD, INFRA_SERVICE_ZK}
	// the infra service endpoint to use:
	endpoint string
	zkconn   *zk.Conn
	kapi     etcd.KeysAPI
	// the storage target to use:
	starget         string
	STORAGE_TARGETS = [...]string{
		STORAGE_TARGET_TTY,
		STORAGE_TARGET_LOCAL,
		STORAGE_TARGET_S3,
		STORAGE_TARGET_MINIO,
	}
	// the backup and restore manifest to use:
	brf      Burryfest
	ErrNoBFF = errors.New("no manifest found")
	cred     string
	// local scratch base directory:
	based string
	// the snapshot ID:
	snapshotid string
	// number of restored items (znodes or keys):
	numrestored int
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
	flag.StringVarP(&bop, "operation", "o", BURRY_OPERATION_BACKUP, fmt.Sprintf("The operation to carry out.\n\tSupported values are %v", BOPS))
	flag.StringVarP(&isvc, "isvc", "i", INFRA_SERVICE_ZK, fmt.Sprintf("The type of infra service to back up or restore.\n\tSupported values are %v", INFRA_SERVICES))
	flag.StringVarP(&endpoint, "endpoint", "e", "", fmt.Sprintf("The infra service HTTP API endpoint to use.\n\tExample: localhost:8181 for Exhibitor"))
	flag.StringVarP(&starget, "target", "t", STORAGE_TARGET_TTY, fmt.Sprintf("The storage target to use.\n\tSupported values are %v", sst))
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
	if snapshotid == "" { // for backup ops
		snapshotid = based
	} else { // for restore ops
		based = snapshotid
	}
	numrestored = 0
}

func processop() bool {
	success := false
	switch bop {
	case BURRY_OPERATION_BACKUP:
		switch brf.InfraService {
		case INFRA_SERVICE_ZK:
			success = backupZK()
		case INFRA_SERVICE_ETCD:
			success = backupETCD()
		default:
			log.WithFields(log.Fields{"func": "process"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	case BURRY_OPERATION_RESTORE:
		switch brf.InfraService {
		case INFRA_SERVICE_ZK:
			success = restoreZK()
		case INFRA_SERVICE_ETCD:
			success = restoreETCD()
		default:
			log.WithFields(log.Fields{"func": "process"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	default:
		log.WithFields(log.Fields{"func": "process"}).Error(fmt.Sprintf("%s is not a valid operation", bop))
		flag.Usage()
		os.Exit(2)
	}
	return success
}

func main() {
	if version {
		about()
		os.Exit(0)
	}
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("Selected operation: %s", strings.ToUpper(bop)))
	log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("My config: %+v", brf))

	if ok := processop(); ok {
		if err := writebf(); err != nil {
			log.WithFields(log.Fields{"func": "main"}).Fatal(fmt.Sprintf("Something went wrong when I tried to create the burry manifest file: %s ", err))
		}
		switch bop {
		case BURRY_OPERATION_BACKUP:
			log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Operation successfully completed. The snapshot ID is: %s", snapshotid))
		case BURRY_OPERATION_RESTORE:
			log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Operation successfully completed. Restored %d items from snapshot %s", numrestored, snapshotid))
		}
	} else {
		log.WithFields(log.Fields{"func": "main"}).Error(fmt.Sprintf("Operation completed with error(s)."))
		flag.Usage()
		os.Exit(1)
	}
}
