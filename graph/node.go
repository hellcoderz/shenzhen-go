// Copyright 2016 Google Inc.
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

package graph

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"unicode"

	"github.com/google/shenzhen-go/parts"
)

// Part abstracts the implementation of a node. Concrete implementations should be
// able to be marshalled to and unmarshalled from JSON sensibly.
type Part interface {
	// AssociateEditor associates a template called "part_view" with the given template.
	AssociateEditor(*template.Template) error

	// Args returns any channel arguments to the part.
	// inputs and outputs map argument names to types (the "<-chan" /
	// "chan<-" part of the type is implied).
	Args() (inputs, outputs map[string]string)

	// Clone returns a copy of this part.
	Clone() interface{}

	// Help returns a helpful description of what this part can do.
	Help() template.HTML

	// Impl returns Go source code implementing the part.
	// The head is executed, then the body is executed (# Multiplicity
	// instances of the body concurrently), then the tail (once the body/bodies
	// are finished).
	//
	// This allows cleanly closing channels for nodes with Multiplicity > 1.
	// The tail is deferred so that the body can use "return" and it is still
	// executed.
	Impl() (head, body, tail string)

	// Imports returns any extra import lines needed for the Part.
	Imports() []string

	// TypeKey returns the "type" of part.
	TypeKey() string

	// Update sets fields in the part based on info in the given Request.
	Update(*http.Request) error
}

// PartFactory creates a part.
type PartFactory func() Part

// PartFactories translates part type strings into part factories.
var PartFactories = map[string]PartFactory{
	/*	"Aggregator":     func() Part { return new(parts.Aggregator) },
		"Broadcast":      func() Part { return new(parts.Broadcast) }, */
	"Code": func() Part { return new(parts.Code) },
	/*	"Filter":         func() Part { return new(parts.Filter) },
		"HTTPServer":     func() Part { return new(parts.HTTPServer) },
		"StaticSend":     func() Part { return new(parts.StaticSend) },
		"TextFileReader": func() Part { return new(parts.TextFileReader) },
		"Unslicer":       func() Part { return new(parts.Unslicer) },*/
}

// Node models a goroutine. It can be marshalled and unmarshalled to JSON sensibly.
type Node struct {
	Part
	Name         string
	Multiplicity uint
	Wait         bool

	// maps arg names to values (channel names)
	params map[string]string
}

// Copy returns a copy of this node, but with an empty name and a clone of the Part.
func (n *Node) Copy() *Node {
	return &Node{
		Name:         "",
		Multiplicity: n.Multiplicity,
		Wait:         n.Wait,
		Part:         n.Part.Clone().(Part),
	}
}

// ImplHead returns the Head part of the implementation.
func (n *Node) ImplHead() string {
	h, _, _ := n.Part.Impl()
	return h
}

// ImplBody returns the Body part of the implementation.
func (n *Node) ImplBody() string {
	_, b, _ := n.Part.Impl()
	return b
}

// ImplTail returns the Tail part of the implementation.
func (n *Node) ImplTail() string {
	_, _, t := n.Part.Impl()
	return t
}

// InputArgs returns the Input args.
func (n *Node) InputArgs() map[string]string {
	i, _ := n.Part.Args()
	return i
}

// OutputArgs returns the Output args.
func (n *Node) OutputArgs() map[string]string {
	_, o := n.Part.Args()
	return o
}

// Args formats the function arguments from the Part as a string.
func (n *Node) Args() string {
	i, o := n.Part.Args()
	a := make([]string, 0, len(i)+len(o))
	ks := make([]string, 0, len(i))
	for k := range i {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		a = append(a, fmt.Sprintf("%s <-chan %s", k, i[k]))
	}
	ks = make([]string, 0, len(o))
	for k := range o {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		a = append(a, fmt.Sprintf("%s chan<- %s", k, o[k]))
	}
	return strings.Join(a, ", ")
}

// Params returns the comma-separated list of channels passed into the function at runtime.
// Only valid after calling mapConnections on the graph.
func (n *Node) Params() string {
	// TODO: Make this less convoluted by storing arg+params together in a slice.
	i, o := n.Part.Args()
	a := make([]string, 0, len(n.params))
	ks := make([]string, 0, len(i))
	for k := range i {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		a = append(a, n.params[k])
	}
	ks = make([]string, 0, len(o))
	for k := range o {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		a = append(a, n.params[k])
	}
	return strings.Join(a, ", ")
}

// Identifier turns the name into a similar-looking identifier.
func (n *Node) Identifier() string {
	base := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		if !unicode.IsLetter(r) && r != '_' && !unicode.IsDigit(r) {
			return -1
		}
		return r
	}, n.Name)
	var f rune
	for _, r := range base {
		f = r
		break
	}
	if unicode.IsDigit(f) {
		base = "_" + base
	}
	return base
}

func (n *Node) String() string { return n.Name }

type jsonNode struct {
	Name         string          `json:"name"`
	Wait         bool            `json:"wait"`
	Multiplicity uint            `json:"multiplicity"`
	Part         json.RawMessage `json:"part"`
	PartType     string          `json:"part_type"`
}

// MarshalJSON encodes the node and part as JSON.
func (n *Node) MarshalJSON() ([]byte, error) {
	p, err := json.Marshal(n.Part)
	if err != nil {
		return nil, err
	}
	if n.Multiplicity < 1 {
		n.Multiplicity = 1
	}
	return json.Marshal(&jsonNode{
		Part:         p,
		PartType:     n.Part.TypeKey(),
		Name:         n.Name,
		Wait:         n.Wait,
		Multiplicity: n.Multiplicity,
	})
}

// UnmarshalJSON decodes the node and part as JSON.
func (n *Node) UnmarshalJSON(j []byte) error {
	var mp jsonNode
	if err := json.Unmarshal(j, &mp); err != nil {
		return err
	}
	pf, ok := PartFactories[mp.PartType]
	if !ok {
		return fmt.Errorf("unknown part type %q", mp.PartType)
	}
	p := pf()
	if err := json.Unmarshal(mp.Part, p); err != nil {
		return err
	}
	if mp.Multiplicity < 1 {
		mp.Multiplicity = 1
	}
	n.Name = mp.Name
	n.Wait = mp.Wait
	n.Multiplicity = mp.Multiplicity
	n.Part = p
	return nil
}
