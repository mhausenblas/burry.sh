package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Burryfest defines the top-level structure of the
// burry manifest file (.burryfest)
type Burryfest struct {
	InfraService  string      `json:"svc"`
	Endpoint      string      `json:"svc-endpoint"`
	StorageTarget string      `json:"target"`
	Creds         Credentials `json:"credentials"`
}

// Credentials defines the structure of the access
// credentials for the storage target endpoint to be used.
type Credentials struct {
	StorageTargetEndpoint string       `json:"target-endpoint"`
	Params                []CredParams `json:"params"`
}

// CredParams defines a generic key-value structure capturing
// credentials for access the storage target endpoint.
type CredParams struct {
	Key   string `json:"key"`
	Value string `json:"val"`
}

// ArchMeta defines the top-level structure for the
// metadata file in the archive.
type ArchMeta struct {
	SnapshotDate          string `json:"snapshot-date"`
	InfraService          string `json:"svc"`
	Endpoint              string `json:"svc-endpoint"`
	StorageTarget         string `json:"target"`
	StorageTargetEndpoint string `json:"target-endpoint"`
}

// S3Config defines the structure for the configuration
// pertaining to the S3 endpoint
type S3Config struct {
	AccessKeyId     string
	SecretAccessKey string
	Bucket          string
	Prefix          string
}

// parsecred parses the cred string in the form:
// STORAGE_TARGET_ENDPOINT,KEY1=VAL1,KEY2=VAL2,...KEYn=VALn
// into a Credentials variable
func parsecred() Credentials {
	c := Credentials{}
	if cred == "" {
		c.Params = []CredParams{}
		return c
	}
	raw := strings.Split(cred, ",")
	params := []CredParams{}
	// 2nd to end are cred params in key-value format:
	for _, p := range raw[1:] {
		p := CredParams{
			Key:   strings.Split(p, "=")[0],
			Value: strings.Split(p, "=")[1],
		}
		params = append(params, p)
	}
	c.StorageTargetEndpoint = raw[0]
	c.Params = params
	return c
}

// extractS3config tries to extract AWS access key and secret
// from an already parsed cred string
func extractS3config() (s3Config S3Config) {
	for _, p := range brf.Creds.Params {
		if p.Key == "ACCESS_KEY_ID" {
			s3Config.AccessKeyId = p.Value
		}
		if p.Key == "SECRET_ACCESS_KEY" {
			s3Config.SecretAccessKey = p.Value
		}
		if p.Key == "BUCKET" {
			s3Config.Bucket = p.Value
		}
		if p.Key == "PREFIX" {
			s3Config.Prefix = p.Value
		}
	}
	return s3Config
}

// loadbf tries to load a JSON representation of the burry manifest
// file from the current working dir.
func loadbf() (string, Burryfest, error) {
	brf = Burryfest{}
	cwd, _ := os.Getwd()
	bfpath, _ := filepath.Abs(filepath.Join(cwd, BURRYFEST_FILE))
	if _, err := os.Stat(bfpath); err != nil { // burryfest does not exist
		return bfpath, brf, err
	} else {
		if raw, ferr := ioutil.ReadFile(bfpath); ferr != nil { // can't read from burryfest
			return bfpath, brf, ferr
		} else {
			if derr := json.Unmarshal(raw, &brf); derr != nil { // can't de-serialize burryfest
				return bfpath, brf, derr
			}
		}
	}
	return bfpath, brf, nil
}

// writebf creates a JSON representation of the burry manifest
// file in the current working dir if and only if such a file
// does not exist, yet.
func writebf() error {
	cwd, _ := os.Getwd()
	bfpath, _ := filepath.Abs(filepath.Join(cwd, BURRYFEST_FILE))
	if _, err := os.Stat(bfpath); err == nil { // burryfest exists, bail out
		return nil
	} else { // burryfest does not exist yet, init it:
		log.WithFields(log.Fields{"func": "writebf"}).Debug(fmt.Sprintf("With credentials %s", brf.Creds))
		if b, err := json.Marshal(brf); err != nil {
			return err
		} else {
			f, err := os.Create(bfpath)
			if err != nil {
				return err
			}
			_, err = f.WriteString(string(b))
			if err != nil {
				return err
			}
			f.Sync()
			log.WithFields(log.Fields{"func": "writebf"}).Debug(fmt.Sprintf("Created burry manifest file %s", bfpath))
			return nil
		}
	}
}

// addmeta adds metadata to the archive.
func addmeta(dst string) error {
	mpath, _ := filepath.Abs(filepath.Join(dst, BURRYMETA_FILE))
	basedi64, _ := strconv.ParseInt(based, 10, 64)
	step := brf.Creds.StorageTargetEndpoint
	if step == "" {
		step, _ = os.Getwd()
	}
	m := ArchMeta{
		SnapshotDate:          time.Unix(basedi64, 0).Format(time.RFC3339),
		InfraService:          brf.InfraService,
		Endpoint:              brf.Endpoint,
		StorageTarget:         brf.StorageTarget,
		StorageTargetEndpoint: step,
	}
	if b, err := json.Marshal(m); err != nil {
		return err
	} else {
		f, err := os.Create(mpath)
		if err != nil {
			return err
		}
		_, err = f.WriteString(string(b))
		if err != nil {
			return err
		}
		f.Sync()
		log.WithFields(log.Fields{"func": "addmeta"}).Debug(fmt.Sprintf("Added metadata to %s", dst))
		return nil
	}
}
