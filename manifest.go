package main

import (
// "encoding/json"
// "os"
// "strings"
)

type BRManifest struct {
	InfraService  string `json:"svc"`
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
