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
	"flag"
	"strings"
	"time"

	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/raft"
	"log"
)

func main() {
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	pqport := flag.Int("port", 9121, "priority queue server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	commitC, errorC, rc := newRaftNode(*id, strings.Split(*cluster, ","), *join, proposeC, confChangeC)


	// Periodic Tick loop
	go func() {
		for {
			select {
			case <- time.After(1 * time.Second):
				if rc.node != nil {
					status := rc.node.Status()
					if status.SoftState.RaftState == raft.StateLeader {
						log.Printf("I AM THE LEADER.")
						log.Printf("Current Leader %v, my id: %v", status.SoftState.Lead, status.ID)
						log.Printf("Basically I can propose the timeout based operations. which will be conceptually similar to used asking us to" +
							"remove it from the delay queue at regular intervals.")
					}

				}

			//do something
			}
		}

	}()

	serveHttpPQueueApi(*pqport, proposeC, confChangeC, commitC, errorC)



	// the key-value http handler will propose updates to raft
	// serveHttpKVAPI(*kvport, proposeC, confChangeC, commitC, errorC)
}
