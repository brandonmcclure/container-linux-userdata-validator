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
		port      int
		address   string
		checkFile string
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
	flag.StringVar(&flags.checkFile, "check-file", "", "only check this file and print report\nThe path can be /dev/stdin for reading from standard input.\nSupported formats are Ignition JSON, coreos-cloudinit config, or a shell script\nA Container Linux Config YAML should be transpiled with ct first.")

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

	if flags.checkFile != "" {
		config, err := ioutil.ReadFile(flags.checkFile)
		if err != nil {
			log.Fatalln(err)
		}
		reports, err := validateInput(config)
		if reports != nil {
			if len(reports) == 0 {
				fmt.Println("Valid user-data")
			} else {
				fmt.Println("Found warnings and errors:")
				for _, entry := range reports {
					fmt.Printf("%v\n", entry)
				}
			}
		}
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

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

func wrapValidateEntries(entries []validate.Entry) []interface{} {
	var interfaceSlice []interface{} = make([]interface{}, len(entries))
	for i, e := range entries {
		interfaceSlice[i] = e
	}
	return interfaceSlice
}

func wrapReportEntries(entries []report.Entry) []interface{} {
	var interfaceSlice []interface{} = make([]interface{}, len(entries))
	for i, e := range entries {
		interfaceSlice[i] = e
	}
	return interfaceSlice
}

func validateInput(config []byte) ([]interface{}, error) {
	_, rpt, err := ignConfig.Parse(config)
	switch err {
	case errors.ErrCloudConfig, errors.ErrEmpty, errors.ErrScript:
		rpt, err := validate.Validate(config)
		if err != nil {
			return nil, err
		}
		return wrapValidateEntries(rpt.Entries()), nil
	case errors.ErrUnknownVersion:
		return wrapReportEntries([]report.Entry{{
			Kind:    report.EntryError,
			Message: "Failed to parse config. Is this a valid Ignition Config, Cloud-Config, or script?",
		}}), nil
	default:
		rpt.Sort()
		return wrapReportEntries(rpt.Entries), nil
	}
}

func putValidate(r *http.Request) (interface{}, nap.Status) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nap.InternalError{err.Error()}
	}

	config := bytes.Replace(body, []byte("\r"), []byte{}, -1)

	reports, err := validateInput(config)
	if err != nil {
		return reports, nap.InternalError{err.Error()}
	}
	return reports, nap.OK{}
}

func getHealth(r *http.Request) (interface{}, nap.Status) {
	return nil, nap.OK{}
}
