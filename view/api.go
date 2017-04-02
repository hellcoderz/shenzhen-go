// Copyright 2017 Google Inc.
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

package view

import (
	"github.com/google/shenzhen-go/controller"
	"log"
	"net/http"
)

type apiHandler struct{}

// API handles API requests.
var API apiHandler

func (apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET api: %v", r.URL.Path)

	g := loadedGraphs[r.URL.Path]
	if g == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := controller.WriteJSONTo(w, g); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Writing JSON: %v", err)
	}
}
