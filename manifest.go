package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
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

// loadbf tries to load a JSON representation of the burry manifest
// file from the current working dir.
func loadbf() (error, string, Burryfest) {
	brf = Burryfest{}
	cwd, _ := os.Getwd()
	bfpath, _ := filepath.Abs(filepath.Join(cwd, BURRYFEST_FILE))
	if _, err := os.Stat(bfpath); err == nil { // a burryfest exists
		if raw, ferr := ioutil.ReadFile(bfpath); ferr != nil {
			return ferr, bfpath, brf
		} else {
			if merr := json.Unmarshal(raw, &brf); merr != nil {
				return merr, bfpath, brf
			}
		}
	} else {
		return ErrNoBFF, bfpath, brf
	}
	return nil, bfpath, brf
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
			log.WithFields(log.Fields{"func": "writebf"}).Info(fmt.Sprintf("Created burry manifest file %s", bfpath))
			return nil
		}
	}
}

// addbf adds the current burry manifest file to the archive. note that this function and copyFileContents
// are derived from http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang/
func addbf(dst string) (err error) {
	cwd, _ := os.Getwd()
	src, _ := filepath.Abs(filepath.Join(cwd, BURRYFEST_FILE))
	log.WithFields(log.Fields{"func": "addbf"}).Info(fmt.Sprintf("Adding %s to %s", src, dst))
	return copyFileContents(src, filepath.Join(dst, BURRYFEST_FILE))
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
