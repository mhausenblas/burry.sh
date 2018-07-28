package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/credentials"
	log "github.com/sirupsen/logrus"
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
	s3Config := extractS3config()
	useSSL := true
	_, f := filepath.Split(localarch)
	if s3Config.Bucket == "" {
		s3Config.Bucket = brf.InfraService + "-backup"
	}
	object := strings.TrimSuffix(f, filepath.Ext(f))
	if len(s3Config.Prefix) != 0 {
		object = filepath.Join(s3Config.Prefix, object)
	}

	log.WithFields(log.Fields{"func": "toremoteS3"}).Debug(fmt.Sprintf("Trying to back up to %s/%s in S3 compatible remote storage", s3Config.Bucket, object))
	mcOpts := minio.Options{}
	mcOpts.Secure = useSSL
	if s3Config.AccessKeyId == "" && s3Config.SecretAccessKey == "" {
		iamCred := credentials.NewIAM("")
		mcOpts.Creds = iamCred
	} else {
		keyCred := credentials.NewStaticV4(s3Config.AccessKeyId, s3Config.SecretAccessKey, "")
		mcOpts.Creds = keyCred
	}
	if mc, err := minio.NewWithOptions(endpoint, &mcOpts); err != nil {
		log.WithFields(log.Fields{"func": "toremoteS3"}).Fatal(fmt.Sprintf("%s ", err))
	} else {
		exists, err := mc.BucketExists(s3Config.Bucket)
		if err != nil || !exists {
			log.WithFields(log.Fields{"func": "toremoteS3"}).Fatal(fmt.Sprintf("%s", err))
		} else {
			if nbytes, err := mc.FPutObject(s3Config.Bucket, object, localarch, minio.PutObjectOptions{
				ContentType: REMOTE_ARCH_TYPE,
			}); err != nil {
				log.WithFields(log.Fields{"func": "toremoteS3"}).Fatal(fmt.Sprintf("%s", err))
			} else {
				log.WithFields(log.Fields{"func": "toremoteS3"}).Info(fmt.Sprintf("Successfully stored %s/%s (%d Bytes) in S3 compatible remote storage %s", s3Config.Bucket, object, nbytes, endpoint))
			}
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
	s3Config := extractS3config()
	useSSL := true
	if s3Config.Bucket == "" {
		s3Config.Bucket = brf.InfraService + "-backup"
	}
	object := snapshotid
	if len(s3Config.Prefix) != 0 {
		object = filepath.Join(s3Config.Prefix, object)
	}

	log.WithFields(log.Fields{"func": "fromremoteS3"}).Debug(fmt.Sprintf("Trying to retrieve %s/%s from S3 compatible remote storage", s3Config.Bucket, object))
	mcOpts := minio.Options{}
	mcOpts.Secure = useSSL
	if s3Config.AccessKeyId == "" && s3Config.SecretAccessKey == "" {
		iamCred := credentials.NewIAM("")
		mcOpts.Creds = iamCred
	} else {
		keyCred := credentials.NewStaticV4(s3Config.AccessKeyId, s3Config.SecretAccessKey, "")
		mcOpts.Creds = keyCred
	}
	if mc, err := minio.NewWithOptions(endpoint, &mcOpts); err != nil {
		log.WithFields(log.Fields{"func": "fromremoteS3"}).Fatal(fmt.Sprintf("%s ", err))
	} else {
		exists, err := mc.BucketExists(s3Config.Bucket)
		if err != nil || !exists {
			log.WithFields(log.Fields{"func": "fromremoteS3"}).Fatal(fmt.Sprintf("%s", err))
		} else {
			if err := mc.FGetObject(s3Config.Bucket, object, localarch, minio.GetObjectOptions{}); err != nil {
				log.WithFields(log.Fields{"func": "fromremoteS3"}).Fatal(fmt.Sprintf("%s", err))
			} else {
				log.WithFields(log.Fields{"func": "fromremoteS3"}).Info(fmt.Sprintf("Successfully retrieved %s/%s from S3 compatible remote storage %s", s3Config.Bucket, object, endpoint))
			}
		}
	}
	return localarch
}
