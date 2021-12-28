package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"distributed_bank/bank"
	"distributed_bank/failuredetector"
	"distributed_bank/leaderdetector"
	multipaxos "distributed_bank/newpaxos"

	"github.com/google/go-cmp/cmp"

	//websocket client
	"distributed_bank/golang-websocket-client/pkg/client"
	//websocket server
	"distributed_bank/golang-websocket-client/pkg/server"
)

// sID: The id of the node running this instance of a Paxos proposer.
//
// nrOfNodes: The total number of Paxos nodes.
//
// adu: all-decided-up-to. The initial id of the highest _consecutive_ slot
// that has been decided. Should normally be set to -1 initially,
func NewDistNetwork(sID int, currentSerList []int, allNodeInfo NodesInfo, aInfo map[int]bank.Account) *DistNetworks {
	adu := -1 //set to -1 initially
	delay := time.Second


	serverAddrmap := make(map[int]string)
	for _, v := range currentSerList {
		serverAddrmap[v] = allNodeInfo.ServerAddrmap[v]
	}

	NI := NodesInfo{ServerIDlist: currentSerList,
		ServerAddrmap: serverAddrmap,
	}


	hbsend := make(chan failuredetector.Heartbeat)
	learnOut := make(chan multipaxos.Learn)
	promiseOut := make(chan multipaxos.Promise)
	prepareOut := make(chan multipaxos.Prepare)
	acceptOut := make(chan multipaxos.Accept)
	decidedOut := make(chan multipaxos.DecidedValue)

	sIDL := NI.ServerIDlist
	nrOfNodes := len(NI.ServerAddrmap)

	leadDect := leaderdetector.NewMonLeaderDetector(sIDL)
	failDect := failuredetector.NewEvtFailureDetector(sID, sIDL, leadDect, delay, hbsend)
	proposer := multipaxos.NewProposer(int(sID), nrOfNodes, adu, leadDect, prepareOut, acceptOut)
	acceptor := multipaxos.NewAcceptor(int(sID), promiseOut, learnOut)
	learner := multipaxos.NewLearner(int(sID), nrOfNodes, decidedOut)
	ldSubscribe := leadDect.Subscribe()

	broadcast:= make(chan server.Message)
	msgIn:=make(chan server.Message) 
	hub := server.NewHub(broadcast,msgIn)

	return &DistNetworks{
		SID: sID,
		Adu:           adu,
		Delay:         time.Second,
		AllNodeInfo:   allNodeInfo,
		CurrentServ:   currentSerList,
		NdInfo:        NI,
		ClientAddrmap: make(map[string]string),
		dcVmap:        make(map[int]multipaxos.DecidedValue),
		AccountMap:    aInfo,

		Ld:       *leadDect,
		Fd:       *failDect,
		Proposer: *proposer,
		Acceptor: *acceptor,
		Learner:  *learner,

		//channels output
		Hbout:       hbsend,
		LearnOut:    learnOut,
		PromiseOut:  promiseOut,
		PrepareOut:  prepareOut,
		AcceptOut:   acceptOut,
		DecidedOut:  decidedOut,
		LdSubscribe: ldSubscribe,

		//parameters for reconfig
		Quorum:     (len(currentSerList) / 2) + 1,
		TimeReconf: int(time.Now().Unix()),
		CprmMap:    make(map[int][]CPromise),
		DcStatMap:  make(map[int]State),

		//channels for reconfig
		ReconfIn: make(chan Reconf, 8),
		NewconfIn: make(chan Newconf, 8),
		CPromiseIn: make(chan CPromise, 8),
		ActIn: make(chan Activiate, 8),
	
		CheckPoint: make(map[int]multipaxos.DecidedValue),
		ClientWSMap: make(map[int]*client.WebSocketClient),
		Broadcast: broadcast,
		MsgIn:msgIn,
		Hub:hub,
	}

}

func (d *DistNetworks) start() {
	d.Fd.Start()
	d.Proposer.Start()
	d.Acceptor.Start()
	d.Learner.Start()
	log.Println("***start()*********the Fd&paxos have been started")
}

func  (d *DistNetworks)StartServerWS(){
	log.Println("***d.CurrentServ",d.CurrentServ)
	for _, v := range d.CurrentServ {
		if d.SID == v {
			log.Println("************the Fd&paxos have been started")
			d.start()
		}
	}

	
	defer d.Hub.CloseConn()//close all the client connections
	go d.Hub.Run()
	//set route for react clients and other servers
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("got new connection from clients and other servers")
		server.ServeWs(d.Hub, w, r)//handles websocket requests from the react clients.
	})

	log.Println("server started ... ")
	addr := d.AllNodeInfo.ServerAddrmap[d.SID]
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Println("err in StartServerWS",err)
		panic(err)
	}

}

//dial all other current severs to send the messages
func (d *DistNetworks) startClientWS(){

	for _,v := range d.CurrentServ{	
		if v!=d.SID{
			log.Println("****startClientWS()",v,d.SID)
			addr := d.AllNodeInfo.ServerAddrmap[v]
			clientws, err := client.NewWebSocketClient(addr, "ws")
			if err != nil {
				panic(err)
			}
			d.ClientWSMap[v]=clientws
		}

	}	
	log.Println("Connecting")
}

//send messages to a single serverWS specified
func (d *DistNetworks) sendMsgToServerWS(msg server.Message, sID int ){
	//if no connection with the specified server, dial the server
	if _, ok := d.ClientWSMap[sID]; !ok {
		addr := d.AllNodeInfo.ServerAddrmap[sID]
		clientws, err := client.NewWebSocketClient(addr, "ws")
		if err != nil {
			panic(err)
		}
		d.ClientWSMap[sID]=clientws
	}
		
	err:=d.ClientWSMap[sID].Write(msg)
	if err != nil {
		log.Println("error in sendMsgToServerWS",err)
		//panic(err)
	}
	

}

//broadcast msg to all current web servers
func (d *DistNetworks) broadcastToServerWS(msg server.Message){
	
			for _, v := range d.CurrentServ {
				if v != d.SID {
					d.ClientWSMap[v].Write(msg)
				}
			}
		
	
}

//send messages to a single react clients
func (d *DistNetworks) sendMsgToReactClient(msg server.Message, cID string){
	d.Hub.DeliverToSend(msg, cID)

	
}
//broadcast to all react clients
func (d *DistNetworks) broadcastToReactClient(msg server.Message){
	d.Hub.DeliverToBC(msg)
}

//deplexing the incoming data from the network
func (d *DistNetworks) handleMessage(msg server.Message) *server.Message {
	switch msg.Command {

	case "CLIENT":
		log.Println("-----received CLIENT msg", msg)
		params := msg.Parameter.(map[string]interface{})
		cID := params["ClientID"].(string)
	
		if d.SID!= d.Ld.Leader() {
			//send LEADER address to the client to inform the leader address
			msg := server.Message{Command: "LEADER", Parameter: d.NdInfo.ServerAddrmap[d.Ld.Leader()]}
			go d.sendMsgToReactClient(msg,cID)
		} 
		break

	case "HEARTBEAT":
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		to := int(params["To"].(float64))
		request := params["Request"].(bool)
		//extract the information from the params and change to the heartbeat format
		hb := failuredetector.Heartbeat{From: from, To: to, Request: request}
		d.Fd.DeliverHeartbeat(hb)
		break
	case "VALUE":
		//A client should be redirected if it connect
		//or send a request to a non-leader node.
		//extract the information from the params and change to the required format
		log.Println("Received VALUE msg", msg)
		params := msg.Parameter.(map[string]interface{})
		cID := params["ClientID"].(string)

		cSq := int(params["ClientSeq"].(float64))
		log.Println("cSq",cSq)
		noop := params["Noop"].(bool)

		accNo := int(params["AccountNum"].(float64))
		log.Println("accNo ",accNo )

		paramsTxn := params["Txn"].(map[string]interface{})
		op := bank.Operation(paramsTxn["Op"].(float64))
		log.Println("op",op)
		amount := int(paramsTxn["Amount"].(float64))
		log.Println("Amount",amount)
		txn := bank.Transaction{Op: op, Amount: amount}
		val := multipaxos.Value{ClientID: cID, ClientSeq: cSq, Noop: noop, AccountNum: accNo, Txn: txn}
		log.Println(" VALUE translated ", val)
		//receive the value from the client, if not leader,
		if d.SID != d.Ld.Leader() {
			//forwards the value request to the leader
			log.Println("Forwards the value to leader...")
			msgRecstrut := server.Message{Command: "VALUE", Parameter: val}
			go d.sendMsgToServerWS(msgRecstrut,d.Ld.Leader())
			//send leader message to the client to inform the leader address
			msgLeader := server.Message{Command: "LEADER", Parameter: d.NdInfo.ServerAddrmap[d.Ld.Leader()]}
			go d.sendMsgToReactClient(msgLeader, cID)
		} else {
			d.Proposer.DeliverClientValue(val)
		}
		break

	case "PROMISE":
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		to := int(params["To"].(float64))
		rnd := multipaxos.Round(params["Rnd"].(float64))
		var slot []multipaxos.PromiseSlot
		if params["Slots"] == nil {
			slot = nil
		} else {
			for _, item := range params["Slots"].([]interface{}) {
				paramsSlots := item.(map[string]interface{})
				id := multipaxos.SlotID(paramsSlots["ID"].(float64))
				vrnd := multipaxos.Round(paramsSlots["Vrnd"].(float64))
				paramsVal := paramsSlots["Vval"].(map[string]interface{})
				vval:=TransValue(paramsVal)
				prmslot := multipaxos.PromiseSlot{ID: id, Vrnd: vrnd, Vval: vval}
				slot = append(slot, prmslot)
			}
		}

		prm := multipaxos.Promise{From: from, To: to, Rnd: rnd, Slots: slot}
		d.Proposer.DeliverPromise(prm)
		break
	case "PREPARE":
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		crnd := multipaxos.Round(params["Crnd"].(float64))
		slot := multipaxos.SlotID(params["Slot"].(float64))
		prep := multipaxos.Prepare{From: from, Slot: slot, Crnd: crnd}
		d.Acceptor.DeliverPrepare(prep)
		break
	case "ACCEPT":
		log.Println(" ACCEPT  ", msg)
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		slot := multipaxos.SlotID(params["Slot"].(float64))
		rnd := multipaxos.Round(params["Rnd"].(float64))
		paramsVal := params["Val"].(map[string]interface{})
		val:=TransValue(paramsVal)
		acc := multipaxos.Accept{From: from, Slot: slot, Rnd: rnd, Val: val}
		d.Acceptor.DeliverAccept(acc)
		break
	case "LEARN":
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		slot := multipaxos.SlotID(params["Slot"].(float64))
		rnd := multipaxos.Round(params["Rnd"].(float64))
		paramsVal := params["Val"].(map[string]interface{})
		val:=TransValue(paramsVal)
		
		ln := multipaxos.Learn{From: from, Slot: slot, Rnd: rnd, Val: val}
		//log.Println(" LEARN translated ", ln)
		d.Learner.DeliverLearn(ln)
		break

	case "RECONFIG":
		params := msg.Parameter.(map[string]interface{})
		newerServStr := params["NewerServStr"].(string)
		ts := int(params["Timestamp"].(float64))

		reconf := Reconf{
			Timestamp:    ts,
			NewerServStr: newerServStr,
		}
		log.Println("***-------------handle message()")
		log.Println("receive reconf msg", reconf)
		d.ReconfIn <- reconf
		break

	case "NEWCONFIG":
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		paramsStat := params["Stat"].(map[string]interface{})
		ts := int(paramsStat["Timestamp"].(float64))
		oldSer := paramsStat["OlderServer"].([]interface{})
		oSList := []int{}
		for _, item := range oldSer {
			oSList = append(oSList, int(item.(float64)))
		}
		newSer := paramsStat["NewerServer"].([]interface{})
		nSList := []int{}
		for _, item := range newSer {
			nSList = append(nSList, int(item.(float64)))
		}
		newconf := Newconf{
			From: from,
			Stat: State{
				Timestamp:   ts,
				OlderServer: oSList,
				NewerServer: nSList,
			},
		}
		d.NewconfIn <- newconf
		break

	case "CPROMISE":
		params := msg.Parameter.(map[string]interface{})
		to := int(params["To"].(float64))
		paramsStat := params["Stat"].(map[string]interface{})
		ts := int(paramsStat["Timestamp"].(float64))
		oldSer := paramsStat["OlderServer"].([]interface{})
		oSList := []int{}
		for _, item := range oldSer {
			oSList = append(oSList, int(item.(float64)))
		}

		newSer := paramsStat["NewerServer"].([]interface{})
		nSList := []int{}
		for _, item := range newSer {
			nSList = append(nSList, int(item.(float64)))
		}

		adu := int(paramsStat["Adu"].(float64))

		aInfo := make(map[int]bank.Account)
		for _, item := range paramsStat["AccountMap"].(map[string]interface{}) {
			p := item.(map[string]interface{})
			bal := int(p["Balance"].(float64))
			num := int(p["Number"].(float64))
			a := bank.Account{Number: num,
				Balance: bal,
			}
			aInfo[num] = a

		}

		checkPoint := make(map[int]multipaxos.DecidedValue)
		for key, item := range paramsStat["CheckPoint"].(map[string]interface{}) {
			k, _ := strconv.Atoi(key)
			p := item.(map[string]interface{})
			slotId := int(p["SlotID"].(float64))
			paramsVal := p["Value"].(map[string]interface{})
			val:=TransValue(paramsVal)
			dcVal := multipaxos.DecidedValue{
				SlotID: multipaxos.SlotID(slotId),
				Value:  val,
			}
			checkPoint[k] = dcVal
		}

		cPrm := CPromise{
			To: to,
			Stat: State{
				Timestamp:   ts,
				OlderServer: oSList,
				NewerServer: nSList,
				Adu:         adu,
				CheckPoint:  checkPoint,
				AccountMap:  aInfo,
			},
		}
		d.CPromiseIn <- cPrm
		break

	case "ACTIVIATE":
		params := msg.Parameter.(map[string]interface{})
		from := int(params["From"].(float64))
		paramsStat := params["Stat"].(map[string]interface{})
		ts := int(paramsStat["Timestamp"].(float64))
		oldSer := paramsStat["OlderServer"].([]interface{})
		oSList := []int{}
		for _, item := range oldSer {
			oSList = append(oSList, int(item.(float64)))
		}

		newSer := paramsStat["NewerServer"].([]interface{})
		nSList := []int{}
		for _, item := range newSer {
			nSList = append(nSList, int(item.(float64)))
		}

		adu := int(paramsStat["Adu"].(float64))

		aInfo := make(map[int]bank.Account)
		for _, item := range paramsStat["AccountMap"].(map[string]interface{}) {
			p := item.(map[string]interface{})
			bal := int(p["Balance"].(float64))
			num := int(p["Number"].(float64))
			a := bank.Account{Number: num,
				Balance: bal,
			}
			aInfo[num] = a

		}
		
		checkPoint := make(map[int]multipaxos.DecidedValue)
		for key, item := range paramsStat["CheckPoint"].(map[string]interface{}) {
			k, _ := strconv.Atoi(key)
			p := item.(map[string]interface{})
			slotId := int(p["SlotID"].(float64))
			paramsVal := p["Value"].(map[string]interface{})
			val:=TransValue(paramsVal)
			dcVal := multipaxos.DecidedValue{
				SlotID: multipaxos.SlotID(slotId),
				Value:  val,
			}
			checkPoint[k] = dcVal
		}

		actv := Activiate{
			From: from,
			Stat: State{
				Timestamp:   ts,
				OlderServer: oSList,
				NewerServer: nSList,
				Adu:         adu,
				CheckPoint:  checkPoint,
				AccountMap:  aInfo,
			},
		}
		d.ActIn <- actv
		log.Println("receive actviation msg ", actv.Stat.Timestamp)
		break

	}

	return nil
}

//handle all the channel
func (d *DistNetworks) handleChan(){

	for {
		select {
		case msg := <-d.MsgIn:
			d.handleMessage(msg)
		case beat := <-d.Hbout:
			if beat.To == beat.From {
				d.Fd.DeliverHeartbeat(beat)
				continue
			}
			msg := server.Message{Command: "HEARTBEAT", Parameter: beat}
			
			sID := beat.To
			go d.sendMsgToServerWS(msg, sID)

		case sub := <-d.LdSubscribe:
			fmt.Println("handleChan() ***********************New leader: %v\n", sub)
		case pre := <-d.PrepareOut:
			d.Acceptor.DeliverPrepare(pre)
			msg := server.Message{Command: "PREPARE", Parameter: pre}
			
			go d.broadcastToServerWS(msg)

		case prm := <-d.PromiseOut:
			if prm.To==d.SID{
				d.Proposer.DeliverPromise(prm)
			}else{			
				msg := server.Message{Command: "PROMISE", Parameter: prm}
				sID:= prm.To
				go d.sendMsgToServerWS(msg, sID)
			}		

		case acc := <-d.AcceptOut:
			d.Acceptor.DeliverAccept(acc)
			msg := server.Message{Command: "ACCEPT", Parameter: acc}
	
			go d.broadcastToServerWS(msg)

		case ln := <-d.LearnOut:
			d.Learner.DeliverLearn(ln)
			msg := server.Message{Command: "LEARN", Parameter: ln}
			go d.broadcastToServerWS(msg)

		case dcV := <-d.DecidedOut:
			d.handleDecidedValue(dcV)

		//receive reconf msg, if it is leader and the msg has newer state,
		//then send newconfig msg to olderserver
		//if it is not leader, forwards to leader
		case reconf := <-d.ReconfIn:
			log.Println("*******---------case reconf := <-d.ReconfIn")
			if d.SID == d.Ld.Leader() {
				if reconf.Timestamp > d.TimeReconf {
					d.Quorum = len(d.CurrentServ)/2 + 1
					//convert the NewerServStr string to the list
					newerSerList := []int{}
					s := strings.Split(reconf.NewerServStr, ",")
					for _, v := range s {
						vInt, _ := strconv.Atoi(v)
						newerSerList = append(newerSerList, vInt)
					}

					//send newconfig msg to older servers
					newconf := Newconf{
						From: d.SID,
						Stat: State{
							Timestamp:   reconf.Timestamp,
							OlderServer: d.CurrentServ,
							NewerServer: newerSerList,
						},
					}
					d.NewconfIn <- newconf
					msg := server.Message{Command: "NEWCONFIG", Parameter: newconf}
					log.Println("msg", msg)
					go d.broadcastToServerWS(msg)
				}
			} else {
				//forwards reconfig msg to leader
				msg := server.Message{Command: "RECONFIG", Parameter: reconf}
				log.Println("msg", msg)
				sID :=d.Ld.Leader()
				go d.sendMsgToServerWS(msg , sID  )

			}

		case newconf := <-d.NewconfIn:
			//receive newconf msg, send cpromise message
			//which contains the state of the server, to the newconfig msg sender
			d.Fd.Stop()
			d.Proposer.Stop()
			d.Acceptor.Stop()
			d.Learner.Stop()
			log.Println("Fd and paxos had been stopped")
			stat := State{
				Timestamp:   newconf.Stat.Timestamp,
				OlderServer: newconf.Stat.OlderServer,
				NewerServer: newconf.Stat.NewerServer,
				Adu:         d.Adu,
				CheckPoint:  d.CheckPoint,
				AccountMap:  d.AccountMap,
			}
	
			cPrm := CPromise{
				To:   newconf.From,
				Stat: stat,
			}
			msg := server.Message{Command: "CPROMISE", Parameter: cPrm}
			sID:=cPrm.To
			go d.sendMsgToServerWS(msg , sID)

		case cPrm := <-d.CPromiseIn:
			d.handleCPrmIn(cPrm)
			
		case actv := <-d.ActIn:
			if actv.Stat.Timestamp >= d.TimeReconf {
				if actv.Stat.Adu >= d.Adu {
					//update states
					d.TimeReconf = actv.Stat.Timestamp
					d.CurrentServ = actv.Stat.NewerServer
					d.Adu = actv.Stat.Adu
					d.CheckPoint = actv.Stat.CheckPoint
					d.AccountMap = actv.Stat.AccountMap
					//update current server address map NI
					serverAddrmap := make(map[int]string)
					for _, v := range d.CurrentServ {
						serverAddrmap[v] = d.AllNodeInfo.ServerAddrmap[v]
					}
	
					NI := NodesInfo{ServerIDlist: d.CurrentServ,
						ServerAddrmap: serverAddrmap,
					}
					d.NdInfo = NI//store current server nodes info
				
					//renewFd&Ld&paxos
					sIDL := d.NdInfo.ServerIDlist
					nrOfNodes := len(d.NdInfo.ServerAddrmap)
					adu := d.Adu
					

					d.Fd.Reconfig(sIDL)
					d.Ld.Reconfig(sIDL)
					d.Proposer = *multipaxos.NewProposer(d.SID, nrOfNodes, adu, &d.Ld, d.PrepareOut, d.AcceptOut)
					// if not reconfig the learner the reconfiguration seems still works
					d.Learner.Reconfig(nrOfNodes)

					//start paxos and Ld
					d.start()
			
				} else {
					stat := State{
						Timestamp:   actv.Stat.Timestamp,
						OlderServer: actv.Stat.OlderServer,
						NewerServer: actv.Stat.NewerServer,
						Adu:         d.Adu,
						CheckPoint:  d.CheckPoint,
						AccountMap:  d.AccountMap,
					}
			
					cPrm := CPromise{
						To:   actv.From,
						Stat: stat,
					}
					msg := server.Message{Command: "CPROMISE", Parameter: cPrm}
					log.Println("has newer state...send it as cpromise")
					log.Println("msg", cPrm.Stat.Timestamp)
					go d.sendMsgToServerWS(msg, cPrm.To)
				}
			}
		}
	}
}

//handle the incoming CPromise, select the one with the highest Adu
func (d *DistNetworks) handleCPrmIn(cPrm CPromise) {
	//cPrm.Stat.Timestamp>=d.TimeReconf
	//store the Cprm in the map, when it receives more than or equal number of Quorum Cprm
	//it begins to vote for the newest state

	if cPrm.Stat.Timestamp >= d.TimeReconf {
		if d.CprmMap[cPrm.Stat.Timestamp] == nil {
			d.CprmMap[cPrm.Stat.Timestamp] = []CPromise{}
			d.CprmMap[cPrm.Stat.Timestamp] = append(d.CprmMap[cPrm.Stat.Timestamp], cPrm)
		} else {
			d.CprmMap[cPrm.Stat.Timestamp] = append(d.CprmMap[cPrm.Stat.Timestamp], cPrm)
			if len(d.CprmMap[cPrm.Stat.Timestamp]) >= d.Quorum {
				//poll for the highest states and return it
				var stat State
				stat.Adu = -1
				for _, v := range d.CprmMap[cPrm.Stat.Timestamp] {
					if v.Stat.Adu > stat.Adu {
						stat = v.Stat
					}
				}
		
				//store the decided state to the DcStatMap

				//if there is no state elected for the corresponding timestamp
				//or the state is newer than the previously decided state
				if (cmp.Equal(d.DcStatMap[stat.Timestamp], State{})) || stat.Adu > d.DcStatMap[stat.Timestamp].Adu {

					d.DcStatMap[stat.Timestamp] = stat
					//send activate msg
					actv := Activiate{
						From: d.SID,
						Stat: stat,
					}
					msg := server.Message{Command: "ACTIVIATE", Parameter: actv}
					newerSer := actv.Stat.NewerServer
					log.Println("-------gonging to send msg ACTIVIATE", actv.Stat.Timestamp)
					for _, v := range newerSer {
						go d.sendMsgToServerWS(msg, v)
					}
				}
			}
		}
	}
}


func (d *DistNetworks) handleDecidedValue(dcV multipaxos.DecidedValue) {

	log.Println("----handleDecidedValue")
	d.CheckPoint[int(d.Adu)] = dcV
	var res multipaxos.Response
	if int(dcV.SlotID) > d.Adu+1 {
		d.dcVmap[int(dcV.SlotID)] = dcV
		return
	}
	if dcV.Value.Noop == false {
		//process the command from the received value
		a := d.AccountMap[dcV.Value.AccountNum]
		if a == (bank.Account{}) {
			log.Println("No such account exists!")
			res = multipaxos.Response{
				ClientID:  dcV.Value.ClientID,
				ClientSeq: dcV.Value.ClientSeq,
				TxnRes: bank.TransactionResult{
					dcV.Value.AccountNum,
					0,
					fmt.Sprintf("No such account exists!"),
				},
			}

		} else {
			txnResult := a.Process(dcV.Value.Txn)
			log.Println(txnResult.String())
			d.AccountMap[dcV.Value.AccountNum] = bank.Account{
				Number:  dcV.Value.AccountNum,
				Balance: txnResult.Balance,
			}

			res = multipaxos.Response{
				ClientID:  dcV.Value.ClientID,
				ClientSeq: dcV.Value.ClientSeq,
				TxnRes:    txnResult,
			}

		}

		//sometimes, the leader falis to receive the client address,
		//as a result the leader can not send the response to the Clients
		//a better solution is to let all the servers send the response to the clients

		msg := server.Message{Command: "RESPONSE", Parameter: res}
		if d.SID== d.Ld.Leader(){
			go d.broadcastToReactClient(msg)
		}
		

		d.Adu++
		d.Proposer.IncrementAllDecidedUpTo()
	}

	if (d.dcVmap[d.Adu+1] != multipaxos.DecidedValue{}) {
		d.handleDecidedValue(d.dcVmap[d.Adu+1])
	}
}

func TransValue(paramsVal map[string]interface{}) multipaxos.Value {
	cID := paramsVal["ClientID"].(string)
	cSq := int(paramsVal["ClientSeq"].(float64))
	noop := paramsVal["Noop"].(bool)
	accNo := int(paramsVal["AccountNum"].(float64))
	paramsTxn := paramsVal["Txn"].(map[string]interface{})
	op := bank.Operation(paramsTxn["Op"].(float64))
	amount := int(paramsTxn["Amount"].(float64))
	txn := bank.Transaction{Op: op, Amount: amount}
	val := multipaxos.Value{ClientID: cID, ClientSeq: cSq, Noop: noop, AccountNum: accNo, Txn: txn}
	return val
}