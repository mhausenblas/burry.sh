package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func about() {
	fmt.Printf("This is burry in version %s\n", VERSION)
}

func get(url string, payload interface{}) error {
	c := &http.Client{Timeout: 2 * time.Second}
	r, err := c.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(payload)
}

func datadirs() {
	stateurl := strings.Join([]string{"http://", endpoint, INFRA_SVC_EXHIBITOR}, "")
	econfig := &ExhibitorState{}
	if err := get(stateurl, econfig); err != nil {
		log.WithFields(log.Fields{"func": "datadirs"}).Error(fmt.Sprintf("Can't parse response from endpoint: %s", err))
	} else {
		log.WithFields(log.Fields{"func": "datadirs"}).Info(fmt.Sprintf("Config %#v", econfig))
	}
}
