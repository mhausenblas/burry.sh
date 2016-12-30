package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Burryfest struct {
	InfraService  string `json:"svc"`
	Endpoint      string `json:"endpoint"`
	StorageTarget string `json:"target"`
	Credentials   string `json:"credentials"`
}

type ExhibitorState struct {
	ExhibitorConfig ExhibitorConfig `json:"config"`
}

type ExhibitorConfig struct {
	LogDir       string `json:"logIndexDirectory"`
	SnapshotsDir string `json:"zookeeperDataDirectory"`
}

func loadbf() (error, Burryfest) {
	brf = Burryfest{}
	cwd, _ := os.Getwd()
	bfpath, _ := filepath.Abs(filepath.Join(cwd, BURRYFEST_FILE))
	if _, err := os.Stat(bfpath); err == nil { // a burryfest exists
		if raw, ferr := ioutil.ReadFile(bfpath); ferr != nil {
			return ferr, brf
		} else {
			if merr := json.Unmarshal(raw, &brf); merr != nil {
				return merr, brf
			}
		}
	} else {
		return ErrNoBFF, brf
	}
	return nil, brf
}

func writebf() error {
	cwd, _ := os.Getwd()
	bfpath, _ := filepath.Abs(filepath.Join(cwd, BURRYFEST_FILE))
	if _, err := os.Stat(bfpath); err == nil {
		log.WithFields(log.Fields{"func": "manifest"}).Info(fmt.Sprintf("Using existing burry manifest file %s", bfpath))
		return nil
	} else { // burryfest does not exist yet, init it:
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
			log.WithFields(log.Fields{"func": "manifest"}).Info(fmt.Sprintf("Created burry manifest file %s", bfpath))
			return nil
		}
	}
}
