// Copyright 2018 Google Inc.
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

package parts

import (
	"bytes"
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
	"github.com/google/shenzhen-go/dev/source"
)

func init() {
	model.RegisterPartType("HTTPServeMux", "Web", &model.PartType{
		New: func() model.Part {
			return &HTTPServeMux{
				Routes: make(map[string]string),
			}
		},
		Panels: []model.PartPanel{
			{
				Name:   "Routes",
				Editor: `TODO(josh): Implement UI`,
			},
			{
				Name:   "Help",
				Editor: `<div><p>HTTPServeMux is a part which routes requests using a <code>http.ServeMux</code>.</p></div>`,
			},
		},
	})
}

// HTTPServeMux is a part which routes requests using a http.ServeMux.
type HTTPServeMux struct {
	// Routes is a map of patterns to output pin names.
	Routes map[string]string `json:"routes"`
}

// Clone returns a clone of this part.
func (m HTTPServeMux) Clone() model.Part {
	r := make(map[string]string, len(m.Routes))
	for k, v := range m.Routes {
		r[k] = v
	}
	return &HTTPServeMux{Routes: r}
}

// Impl returns the implementation.
func (m HTTPServeMux) Impl(types map[string]string) (head, body, tail string) {
	// I think http.ServeMux is concurrent safe... it guards everything with RWMutex.
	hb, tb := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	closed := source.NewStringSet()

	hb.WriteString("mux := http.NewServeMux()\n")
	for pat, out := range m.Routes {
		fmt.Fprintf(hb, "mux.Handle(%q, parts.HTTPHandler(%s))\n", pat, out)

		if closed.Ni(out) {
			continue
		}
		fmt.Fprintf(tb, "close(%s)\n", out)
		closed.Add(out)
	}

	return hb.String(),
		`for req := range requests {
			h, _ := mux.Handler(req.Request)
			hh, ok := h.(*parts.HTTPHandler)
			if !ok {
				panic("mux contained a http.Handler that wasn't actually*parts.HTTPHandler")
			}
			hh <- req
		}`,
		tb.String()
}

// Imports returns needed imports.
func (m HTTPServeMux) Imports() []string {
	return []string{
		`"net/http"`,
		`"github.com/google/shenzhen-go/dev/parts"`,
	}
}

// Pins returns a pin map, in this case varying by configuration.
func (m HTTPServeMux) Pins() pin.Map {
	p := pin.NewMap(&pin.Definition{
		Name:      "requests",
		Direction: pin.Input,
		Type:      "*parts.HTTPRequest",
	})
	for _, out := range m.Routes {
		if p[out] != nil {
			// Nothing wrong with routing multiple patterns to the same output.
			// Even if it didn't skip here, it would set the same definition...
			continue
		}
		p[out] = &pin.Definition{
			Name:      out,
			Direction: pin.Output,
			Type:      "*parts.HTTPRequest",
		}
	}
	return p
}

// TypeKey returns "HTTPServeMux".
func (m HTTPServeMux) TypeKey() string {
	return "HTTPServeMux"
}
