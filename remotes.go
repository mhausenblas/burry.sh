package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/minio/minio-go"
	"os"
	"path/filepath"
	"strings"
)

// remote uploads the local ZIP archive to a
// remote storage target such as S3 or Minio
func remote(localarch string) {
	stidx := lookupst(brf.StorageTarget)
	switch {
	case stidx == 0, stidx == 1: // either TTY or local storage so we're done
		return
	case stidx == 2, stidx == 3: // S3 compativle remote storage
		remoteS3(localarch)
	default:
		log.WithFields(log.Fields{"func": "remote"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
	}
}

// remoteS3 handles S3 compatible (remote) storage targets
func remoteS3(localarch string) {
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

	log.WithFields(log.Fields{"func": "remoteS3"}).Info(fmt.Sprintf("Trying to back up to %s/%s in S3 compatible remote storage", bucket, object))
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
			log.WithFields(log.Fields{"func": "remoteS3"}).Info(fmt.Sprintf("Successfully stored %s/%s (%d Bytes) in S3 compatible remote storage %s", bucket, object, nbytes, endpoint))
		}
	}
}
