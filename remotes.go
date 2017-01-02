package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/minio/minio-go"
	"os"
	"path/filepath"
	"strings"
)

// toremote uploads the local ZIP archive to a
// remote storage target such as S3 or Minio
func toremote(localarch string) {
	stidx := lookupst(brf.StorageTarget)
	switch {
	case stidx == 0, stidx == 1: // either TTY or local storage so we're done
		return
	case stidx == 2, stidx == 3: // S3 compatible remote storage
		toremoteS3(localarch)
	default:
		log.WithFields(log.Fields{"func": "toremote"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
	}
}

// toremoteS3 handles storing an archive in S3 compatible (remote) storage targets
func toremoteS3(localarch string) {
	defer func() {
		_ = os.Remove(localarch)
	}()
	endpoint := brf.Creds.StorageTargetEndpoint
	accessKeyID := ""
	secretAccessKey := ""
	for _, p := range brf.Creds.Params {
		if p.Key == "AWS_ACCESS_KEY_ID" {
			accessKeyID = p.Value
		}
		if p.Key == "AWS_SECRET_ACCESS_KEY" {
			secretAccessKey = p.Value
		}
	}
	useSSL := true
	_, f := filepath.Split(localarch)
	bucket := brf.InfraService + "-backup-" + strings.TrimSuffix(f, filepath.Ext(f))
	object := "latest.zip"
	ctype := "application/zip"

	log.WithFields(log.Fields{"func": "toremoteS3"}).Info(fmt.Sprintf("Trying to back up to %s/%s in S3 compatible remote storage", bucket, object))
	if mc, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL); err != nil {
		log.WithFields(log.Fields{"func": "toremoteS3"}).Fatal(fmt.Sprintf("%s ", err))
	} else {
		location := "us-east-1"
		if err = mc.MakeBucket(bucket, location); err != nil {
			exists, err := mc.BucketExists(bucket)
			if err == nil && exists {
				log.WithFields(log.Fields{"func": "toremoteS3"}).Info(fmt.Sprintf("Bucket %s already exists", bucket))
			} else {
				log.WithFields(log.Fields{"func": "toremoteS3"}).Fatal(fmt.Sprintf("%s", err))
			}
		}
		if nbytes, err := mc.FPutObject(bucket, object, localarch, ctype); err != nil {
			log.WithFields(log.Fields{"func": "toremoteS3"}).Fatal(fmt.Sprintf("%s", err))
		} else {
			log.WithFields(log.Fields{"func": "toremoteS3"}).Info(fmt.Sprintf("Successfully stored %s/%s (%d Bytes) in S3 compatible remote storage %s", bucket, object, nbytes, endpoint))
		}
	}
}

// fromremote downloads a ZIP archive from a
// remote storage target such as S3 or Minio
func fromremote() string {
	stidx := lookupst(brf.StorageTarget)
	cwd, _ := os.Getwd()
	switch {
	case stidx == 1: // local storage so a NOP, only construct file name
		return filepath.Join(cwd, based+".zip")
	case stidx == 2, stidx == 3: // S3 compatible remote storage
		return fromremoteS3()
	default:
		log.WithFields(log.Fields{"func": "fromremote"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
		return ""
	}
}

// fromremoteS3 handles retrieving an archive from S3 compatible (remote) storage targets
// into the current working directory
func fromremoteS3() string {
	cwd, _ := os.Getwd()
	localarch := filepath.Join(cwd, based+".zip")
	endpoint := brf.Creds.StorageTargetEndpoint
	accessKeyID := ""
	secretAccessKey := ""
	for _, p := range brf.Creds.Params {
		if p.Key == "AWS_ACCESS_KEY_ID" {
			accessKeyID = p.Value
		}
		if p.Key == "AWS_SECRET_ACCESS_KEY" {
			secretAccessKey = p.Value
		}
	}
	useSSL := true
	bucket := brf.InfraService + "-backup-" + snapshotid
	object := "latest.zip"

	log.WithFields(log.Fields{"func": "fromremoteS3"}).Info(fmt.Sprintf("Trying to retrieve %s/%s from S3 compatible remote storage", bucket, object))
	if mc, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL); err != nil {
		log.WithFields(log.Fields{"func": "fromremoteS3"}).Fatal(fmt.Sprintf("%s ", err))
	} else {
		exists, err := mc.BucketExists(bucket)
		if err != nil || !exists {
			log.WithFields(log.Fields{"func": "fromremoteS3"}).Fatal(fmt.Sprintf("%s", err))
		} else {
			if err := mc.FGetObject(bucket, object, localarch); err != nil {
				log.WithFields(log.Fields{"func": "fromremoteS3"}).Fatal(fmt.Sprintf("%s", err))
			} else {
				log.WithFields(log.Fields{"func": "fromremoteS3"}).Info(fmt.Sprintf("Successfully retrieved %s/%s from S3 compatible remote storage %s", bucket, object, endpoint))
			}
		}
	}
	return localarch
}
