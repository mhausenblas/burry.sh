package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/minio/minio-go"
	azip "github.com/pierrre/archivefile/zip"
	"os"
	"path/filepath"
	"strings"
)

func about() {
	fmt.Printf("This is burry in version %s\n", VERSION)
}

func lookupst(name string) int {
	switch strings.ToLower(name) {
	case "tty":
		return 0
	case "local":
		return 1
	case "s3":
		return 2
	default:
		return -1
	}
}

func store(path string, val string) {
	cwd, _ := os.Getwd()
	fpath := ""
	if path == "/" {
		log.WithFields(log.Fields{"func": "store"}).Info(fmt.Sprintf("Rewriting root"))
		fpath, _ = filepath.Abs(filepath.Join(cwd, based))
	} else {
		fpath, _ = filepath.Abs(filepath.Join(cwd, based, strings.Replace(path, ":", "BURRY_ESC_COLON", -1)))
	}
	if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
		log.WithFields(log.Fields{"func": "store"}).Error(fmt.Sprintf("%s", err))
		return
	} else {
		cpath, _ := filepath.Abs(filepath.Join(fpath, "content"))
		if c, cerr := os.Create(cpath); cerr != nil {
			log.WithFields(log.Fields{"func": "store"}).Error(fmt.Sprintf("%s", cerr))
		} else {
			defer c.Close()
			if nbytes, err := c.WriteString(val); err != nil {
				log.WithFields(log.Fields{"func": "store"}).Error(fmt.Sprintf("%s", err))
			} else {
				log.WithFields(log.Fields{"func": "store"}).Debug(fmt.Sprintf("Stored %s in %s with %d bytes", path, fpath, nbytes))
			}
		}
	}
}

func arch() string {
	defer func() {
		_ = os.RemoveAll(based)
	}()
	cwd, _ := os.Getwd()
	opath := filepath.Join(cwd, based+".zip")
	ipath := filepath.Join(cwd, based, "/")
	progress := func(apath string) {
		log.WithFields(log.Fields{"func": "arch"}).Debug(fmt.Sprintf("%s", apath))
	}
	// add the current burry manifest file as metadata:
	addbf(ipath)
	if err := azip.ArchiveFile(ipath, opath, progress); err != nil {
		log.WithFields(log.Fields{"func": "arch"}).Panic(fmt.Sprintf("%s", err))
	} else {
		log.WithFields(log.Fields{"func": "arch"}).Info(fmt.Sprintf("Backup available in %s", opath))
	}
	return opath
}

func remote(localarch string) {
	stidx := lookupst(brf.StorageTarget)
	switch {
	case stidx == 0, stidx == 1: // either TTY or local storage so we're done
		return
	case stidx == 2: // Amazon S3
		remoteS3(localarch)
	default:
		log.WithFields(log.Fields{"func": "remote"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
	}
}

func remoteS3(localarch string) {
	defer func() {
		_ = os.Remove(localarch)
	}()
	endpoint := "play.minio.io:9000"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true
	_, f := filepath.Split(localarch)
	bucket := brf.InfraService + "-backup-" + strings.TrimSuffix(f, filepath.Ext(f))
	object := "latest.zip"
	ctype := "application/zip"

	log.WithFields(log.Fields{"func": "remoteS3"}).Info(fmt.Sprintf("Trying to back up to %s/%s in Amazon S3", bucket, object))
	if mc, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL); err != nil {
		log.WithFields(log.Fields{"func": "remoteS3"}).Fatal(fmt.Sprintf("%s ", err))
	} else {
		location := "us-east-1"
		if err = mc.MakeBucket(bucket, location); err != nil {
			exists, err := mc.BucketExists(bucket)
			if err == nil && exists {
				log.WithFields(log.Fields{"func": "remoteS3"}).Info(fmt.Sprintf("Bucket %s already exists", bucket))
			} else {
				log.WithFields(log.Fields{"func": "remoteS3"}).Fatal(fmt.Sprintf("%s", err))
			}
		}
		if nbytes, err := mc.FPutObject(bucket, object, localarch, ctype); err != nil {
			log.WithFields(log.Fields{"func": "remoteS3"}).Fatal(fmt.Sprintf("%s", err))
		} else {
			log.WithFields(log.Fields{"func": "remoteS3"}).Info(fmt.Sprintf("Successfully stored %s/%s (%d Bytes) in Amazon S3", bucket, object, nbytes))
		}
	}
}
