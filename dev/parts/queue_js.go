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

//+build js

package parts

import "github.com/google/shenzhen-go/dev/dom"

var (
	doc = dom.CurrentDocument()

	inputQueueMaxItems = doc.ElementByID("queue-maxitems")
	selectQueueMode    = doc.ElementByID("queue-mode")

	focusedQueue *Queue
)

func init() {
	inputQueueMaxItems.AddEventListener("change", func(dom.Object) {
		focusedQueue.MaxItems = inputQueueMaxItems.Get("value").Int()
	})
	selectQueueMode.AddEventListener("change", func(dom.Object) {
		focusedQueue.Mode = QueueMode(selectQueueMode.Get("value").String())
	})
}

func (q *Queue) GainFocus() {
	focusedQueue = q
	inputQueueMaxItems.Set("value", q.MaxItems)
	selectQueueMode.Set("value", q.Mode)
}