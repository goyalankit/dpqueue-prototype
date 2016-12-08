// Copyright 2015 CoreOS, Inc.
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
	"encoding/gob"
	"log"
	"strconv"
	"sync"
	"fmt"
)

// a key-value store backed by raft
type DPQueue struct {
	proposeC chan<- string // channel for proposing updates
	mu       sync.RWMutex
	pqueue   *PQueue // current committed key-value pairs
}

func newDPQueue(proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *DPQueue {
	s := &DPQueue{proposeC: proposeC, pqueue: NewPQueue(MAXPQ)}
	// replay log into key-value map
	s.readCommits(commitC, errorC)
	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s
}

// Don't do anything here yet.
func (s *DPQueue) Lookup(key string) (string, bool) {
	s.mu.RLock()
	// v, ok := s.kvStore[key
	value, priority := s.pqueue.Head()
	result := fmt.Sprintf("%r -> %r", value, priority)
	s.mu.RUnlock()
	// return v, ok
	return result, true
}

func (s *DPQueue) Propose(k string, v string) {
	log.Printf("Proposing value: %v - %v", k, v)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data_t{k, v}); err != nil {
		log.Fatal(err)
	}
	s.proposeC <- string(buf.Bytes())
}

type data_t struct {
	Value    string
	Priority string
}

func (s *DPQueue) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			return
		}

		var data_v data_t
		log.Printf("Received data.\n")
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&data_v); err != nil {
			log.Fatalf("raftexample: could not decode message (%v)", err)
		}
		s.mu.Lock()
		//s.kvStore[data_kv.Key] = data_kv.Val
		vv, _ := strconv.Atoi(data_v.Priority)
		s.pqueue.Push(data_v.Value, vv)
		s.mu.Unlock()
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}
