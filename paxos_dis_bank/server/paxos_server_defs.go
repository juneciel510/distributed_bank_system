package main

import (
	"net"
	"sync"
	"time"

	"distributed_bank/bank"
	"distributed_bank/failuredetector"
	"distributed_bank/golang-websocket-client/pkg/client"
	"distributed_bank/golang-websocket-client/pkg/server"
	"distributed_bank/leaderdetector"
	multipaxos "distributed_bank/newpaxos"
)

type State struct {
	Timestamp   int
	OlderServer []int
	NewerServer []int
	Adu         int
	CheckPoint  map[int]multipaxos.DecidedValue //map for storing decided value
	AccountMap  map[int]bank.Account
}

type Reconf struct {
	Timestamp    int
	NewerServStr string
}

type Newconf struct {
	From int
	Stat State
}

type CPromise struct {
	To   int
	Stat State
}

type Activiate struct {
	From int
	Stat State
}

type Reconfresp struct {
	SId   int
	SAddr string
}

type NodesInfo struct {
	ServerIDlist  []int
	ServerAddrmap map[int]string
}

type TcpServer struct {
	Con net.Listener
	SId   int
	SAddr string

}



type ClientInfo struct {
	ClientID   string
	ClientAddr string
}

type SyncClientWSMap struct {
	ClientWSMap map[int]*client.WebSocketClient//store the websocket of all other servers
	ClientWSMapMutex sync.RWMutex
}
type DistNetworks struct {
	SID int
	Adu           int
	Delay         time.Duration
	AllNodeInfo   NodesInfo
	CurrentServ   []int
	NdInfo        NodesInfo//store current server nodes info
	ClientAddrmap map[string]string
	dcVmap        map[int]multipaxos.DecidedValue //key is value.slotID
	AccountMap    map[int]bank.Account            //map for storing account info
	ResponseMap   map[int]multipaxos.Response

	Ld       leaderdetector.MonLeaderDetector
	Fd       failuredetector.EvtFailureDetector
	Proposer multipaxos.Proposer
	Acceptor multipaxos.Acceptor
	Learner  multipaxos.Learner


	Hbout      chan failuredetector.Heartbeat
	PrepareOut chan multipaxos.Prepare
	PromiseOut chan multipaxos.Promise
	AcceptOut  chan multipaxos.Accept
	LearnOut   chan multipaxos.Learn
	DecidedOut chan multipaxos.DecidedValue

	LdSubscribe <-chan int


	Quorum     int
	TimeReconf int //the time of last configuration
	CprmMap    map[int][]CPromise
	DcStatMap  map[int]State

	ReconfIn chan Reconf
	NewconfIn  chan Newconf
	CPromiseIn chan CPromise
	ActIn chan Activiate
	CheckPoint map[int]multipaxos.DecidedValue //key is adu

	// ClientWSMap map[int]*client.WebSocketClient//store the websocket of all other servers
	SyncCWSM SyncClientWSMap
	Hub *server.Hub
	Broadcast chan server.Message
	MsgIn chan server.Message

}
