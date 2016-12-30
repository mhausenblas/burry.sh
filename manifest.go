package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
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

func manifest() error {
	if b, err := json.Marshal(brf); err != nil {
		return err
	} else {
		cwd, _ := os.Getwd()
		bf, _ := filepath.Abs(filepath.Join(cwd, ".burryfest"))
		f, err := os.Create(bf)
		if err != nil {
			return err
		}
		_, err = f.WriteString(string(b))
		if err != nil {
			return err
		}
		f.Sync()
		log.WithFields(log.Fields{"func": "manifest"}).Info(fmt.Sprintf("Created burry manifest file %s", bf))
		return nil
	}
}
