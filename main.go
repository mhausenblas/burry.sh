package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	consul "github.com/hashicorp/consul/api"
	flag "github.com/ogier/pflag"
	"github.com/samuel/go-zookeeper/zk"
)

const (
	VERSION                 string = "0.4.0"
	BURRYFEST_FILE          string = ".burryfest"
	BURRYMETA_FILE          string = ".burrymeta"
	CONTENT_FILE            string = "content"
	BURRY_OPERATION_BACKUP  string = "backup"
	BURRY_OPERATION_RESTORE string = "restore"
	INFRA_SERVICE_ETCD      string = "etcd"
	INFRA_SERVICE_ZK        string = "zk"
	INFRA_SERVICE_CONSUL    string = "consul"
	STORAGE_TARGET_TTY      string = "tty"
	STORAGE_TARGET_LOCAL    string = "local"
	STORAGE_TARGET_S3       string = "s3"
	STORAGE_TARGET_MINIO    string = "minio"
	REMOTE_ARCH_FILE        string = "latest.zip"
	REMOTE_ARCH_TYPE        string = "application/zip"
)

var (
	version         bool
	createburryfest bool
	// the operation burry should to carry out:
	bop  string
	bops = [...]string{BURRY_OPERATION_BACKUP, BURRY_OPERATION_RESTORE}
	// the type of infra service to back up or restore:
	isvc  string
	isvcs = [...]string{INFRA_SERVICE_ZK, INFRA_SERVICE_ETCD, INFRA_SERVICE_CONSUL}
	// the infra service endpoint to use:
	endpoint string
	zkconn   *zk.Conn
	kapi     etcd.KeysAPI
	ckv      *consul.KV
	// the storage target to use:
	starget   string
	startgets = [...]string{
		STORAGE_TARGET_TTY,
		STORAGE_TARGET_LOCAL,
		STORAGE_TARGET_S3,
		STORAGE_TARGET_MINIO,
	}
	// the backup and restore manifest to use:
	brf  Burryfest
	cred string
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
	flag.BoolVarP(&version, "version", "v", false, "Display version information and exit.")
	flag.BoolVarP(&createburryfest, "burryfest", "b", false, fmt.Sprintf("Create a burry manifest file %s in the current directory.\n\tThe manifest file captures the current command line parameters for re-use in subsequent operations.", BURRYFEST_FILE))
	flag.StringVarP(&bop, "operation", "o", BURRY_OPERATION_BACKUP, fmt.Sprintf("The operation to carry out.\n\tSupported values are %v", bops))
	flag.StringVarP(&isvc, "isvc", "i", INFRA_SERVICE_ZK, fmt.Sprintf("The type of infra service to back up or restore.\n\tSupported values are %v", isvcs))
	flag.StringVarP(&endpoint, "endpoint", "e", "", fmt.Sprintf("The infra service HTTP API endpoint to use.\n\tExample: localhost:8181 for Exhibitor"))
	flag.StringVarP(&starget, "target", "t", STORAGE_TARGET_TTY, fmt.Sprintf("The storage target to use.\n\tSupported values are %v", startgets))
	flag.StringVarP(&cred, "credentials", "c", "", fmt.Sprintf("The credentials to use in format STORAGE_TARGET_ENDPOINT,KEY1=VAL1,...KEYn=VALn.\n\tExample: s3.amazonaws.com,ACCESS_KEY_ID=...,SECRET_ACCESS_KEY=..."))
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

	if bfpath, mbrf, err := loadbf(); err != nil {
		brf = Burryfest{InfraService: isvc, Endpoint: endpoint, StorageTarget: starget, Creds: parsecred()}
	} else {
		brf = mbrf
		log.WithFields(log.Fields{"func": "init"}).Info(fmt.Sprintf("Using burryfest %s", bfpath))
	}

	based = strconv.FormatInt(time.Now().Unix(), 10)
	if bop == BURRY_OPERATION_BACKUP {
		snapshotid = based
	} else { // for restore ops
		if snapshotid != "" {
			based = snapshotid
		}
	}
	numrestored = 0
}

func processop() bool {
	success := false
	// validate available operations parameter:
	if brf.Endpoint == "" {
		log.WithFields(log.Fields{"func": "processop"}).Error(fmt.Sprintf("You MUST supply an infra service endpoint with -e/--endpoint"))
		return false
	}
	switch bop {
	case BURRY_OPERATION_BACKUP:
		switch brf.InfraService {
		case INFRA_SERVICE_ZK:
			success = backupZK()
		case INFRA_SERVICE_ETCD:
			success = backupETCD()
		case INFRA_SERVICE_CONSUL:
			success = backupCONSUL()
		default:
			log.WithFields(log.Fields{"func": "processop"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	case BURRY_OPERATION_RESTORE:
		if brf.StorageTarget == STORAGE_TARGET_TTY {
			log.WithFields(log.Fields{"func": "processop"}).Error(fmt.Sprintf("I can't restore from TTY, pick a different storage target with -t/--target"))
			return false
		}
		if snapshotid == "" {
			log.WithFields(log.Fields{"func": "processop"}).Error(fmt.Sprintf("You MUST supply a snapshot ID with -s/--snapshot"))
			return false
		}
		switch brf.InfraService {
		case INFRA_SERVICE_ZK:
			success = restoreZK()
		case INFRA_SERVICE_ETCD:
			success = restoreETCD()
		case INFRA_SERVICE_CONSUL:
			success = restoreCONSUL()
		default:
			log.WithFields(log.Fields{"func": "processop"}).Error(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
	default:
		log.WithFields(log.Fields{"func": "processop"}).Error(fmt.Sprintf("%s is not a valid operation", bop))
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
	log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("Selected operation: %s", strings.ToUpper(bop)))
	log.WithFields(log.Fields{"func": "main"}).Info(fmt.Sprintf("My config: %+v", brf))

	if ok := processop(); ok {
		if createburryfest {
			if err := writebf(); err != nil {
				log.WithFields(log.Fields{"func": "main"}).Fatal(fmt.Sprintf("Something went wrong when I tried to create the burryfest: %s ", err))
			}
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
