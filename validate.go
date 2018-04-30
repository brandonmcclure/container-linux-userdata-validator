//
// Copyright 2015 The CoreOS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/coreos/coreos-cloudinit/config/validate"
	"github.com/coreos/ignition/config/shared/errors"
	ignConfig "github.com/coreos/ignition/config/v2_2"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/crawford/nap"
	"github.com/gorilla/mux"
)

var (
	flags = struct {
		port    int
		address string
	}{}
)

type payloadWrapper struct{}

func (w payloadWrapper) Wrap(payload interface{}, status nap.Status) (interface{}, int) {
	return map[string]interface{}{
		"result": payload,
	}, status.Code()
}

type panicHandler struct{}

func (h panicHandler) Handle(e interface{}) {
	log.Printf("PANIC: %#v\n", e)
	debug.PrintStack()
}

func init() {
	flag.StringVar(&flags.address, "address", "0.0.0.0", "address to listen on")
	flag.IntVar(&flags.port, "port", 80, "port to bind on")

	nap.PayloadWrapper = payloadWrapper{}
	nap.PanicHandler = panicHandler{}
	nap.ResponseHeaders = []nap.Header{
		nap.Header{"Access-Control-Allow-Origin", []string{"*"}},
		nap.Header{"Access-Control-Allow-Methods", []string{"OPTIONS, PUT"}},
		nap.Header{"Content-Type", []string{"application/json"}},
		nap.Header{"Cache-Control", []string{"no-cache,must-revalidate"}},
	}
}

func main() {
	flag.Parse()

	router := mux.NewRouter()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", flags.address, flags.port),
		Handler: router,
	}

	router.Handle("/validate", nap.HandlerFunc(optionsValidate)).Methods("OPTIONS")
	router.Handle("/validate", nap.HandlerFunc(putValidate)).Methods("PUT")
	router.Handle("/health", nap.HandlerFunc(getHealth)).Methods("GET")

	log.Fatalln(server.ListenAndServe())
}

func optionsValidate(r *http.Request) (interface{}, nap.Status) {
	return nil, nap.OK{}
}

func putValidate(r *http.Request) (interface{}, nap.Status) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nap.InternalError{err.Error()}
	}

	config := bytes.Replace(body, []byte("\r"), []byte{}, -1)

	_, rpt, err := ignConfig.Parse(config)
	switch err {
	case errors.ErrCloudConfig, errors.ErrEmpty, errors.ErrScript:
		rpt, err := validate.Validate(config)
		if err != nil {
			return nil, nap.InternalError{err.Error()}
		}
		return rpt.Entries(), nap.OK{}
	case errors.ErrUnknownVersion:
		return []report.Entry{{
			Kind:    report.EntryError,
			Message: "Failed to parse config. Is this a valid Ignition Config, Cloud-Config, or script?",
		}}, nap.OK{}
	default:
		rpt.Sort()
		return rpt.Entries, nap.OK{}
	}
}

func getHealth(r *http.Request) (interface{}, nap.Status) {
	return nil, nap.OK{}
}
