package multipaxos

import (
	"sort"
)

// Acceptor represents an acceptor as defined by the Multi-Paxos algorithm.
type Acceptor struct { 
	id  int   // id: The id of the node running this instance of a Paxos acceptor.
	rnd Round //highest round seen
	historyVal map[SlotID]PromiseSlot //the slotID and its round in which a value last accepted

	learnOut   chan<- Learn
	promiseOut chan<- Promise
	acceptIn   chan Accept
	prepareIn  chan Prepare

	stop chan struct{}
}

// NewAcceptor returns a new Multi-Paxos acceptor.
// It takes the following arguments:
//
// id: The id of the node running this instance of a Paxos acceptor.
//
// promiseOut: A send only channel used to send promises to other nodes.
//
// learnOut: A send only channel used to send learns to other nodes.
func NewAcceptor(id int, promiseOut chan<- Promise, learnOut chan<- Learn) *Acceptor {
	return &Acceptor{
		id:  id,
		rnd: NoRound,

		historyVal: make(map[SlotID]PromiseSlot),

		promiseOut: promiseOut,
		learnOut:   learnOut,
		acceptIn:   make(chan Accept, 8),
		prepareIn:  make(chan Prepare, 8),

		stop: make(chan struct{}),
	}
}

// Start starts a's main run loop as a separate goroutine. The main run loop
// handles incoming prepare and accept messages.
func (a *Acceptor) Start() {
	go func() {
		for {
			select {
			case prp := <-a.prepareIn:
				promises, output := a.handlePrepare(prp)
				if !output {
					continue
				}
				a.promiseOut <- promises
			case acc := <-a.acceptIn:
				learns, output := a.handleAccept(acc)
				if !output {
					continue
				}
				a.learnOut <- learns
			case <-a.stop:
				return
			}
		}
	}()
}

// Stop stops a's main run loop.
func (a *Acceptor) Stop() {
	a.stop <- struct{}{}
}

// DeliverPrepare delivers prepare prp to acceptor a.
func (a *Acceptor) DeliverPrepare(prp Prepare) {
	a.prepareIn <- prp
}

// DeliverAccept delivers accept acc to acceptor a.
func (a *Acceptor) DeliverAccept(acc Accept) {
	a.acceptIn <- acc
}

// Internal: handlePrepare processes prepare prp according to the Multi-Paxos
// algorithm. If handling the prepare results in acceptor a emitting a
// corresponding promise, then output will be true and prm contain the promise.
// If handlePrepare returns false as output, then prm will be a zero-valued
// struct.
func (a *Acceptor) handlePrepare(prp Prepare) (prm Promise, output bool) {
	if prp.Crnd > a.rnd {
		a.rnd = prp.Crnd //update highest seen round

		prm.To = prp.From
		prm.From = a.id
		prm.Rnd = prp.Crnd
		//deliver the history accepted values
		//slotsHistory := make([]PromiseSlot)
		var slotsHistory []PromiseSlot
		maplen := len(a.historyVal)
		if maplen > 0 {

			var index []int
			for k := range a.historyVal {
				if k < prp.Slot {
					delete(a.historyVal, k)
				}
			}
			for k := range a.historyVal {
				index = append(index, int(k))
			}

			sort.Ints(index)
			for _, value := range index { 
				slotsHistory = append(slotsHistory, a.historyVal[SlotID(value)])
			}

		}
		prm.Slots = slotsHistory
		output = true
	} else {
		output = false

	}

	return prm, output
}

// Internal: handleAccept processes accept acc according to the Multi-Paxos
// algorithm. If handling the accept results in acceptor a emitting a
// corresponding learn, then output will be true and lrn contain the learn.  If
// handleAccept returns false as output, then lrn will be a zero-valued struct.
func (a *Acceptor) handleAccept(acc Accept) (lrn Learn, output bool) {
	vrnd := a.historyVal[acc.Slot].Vrnd
	if acc.Rnd >= a.rnd && acc.Rnd != vrnd {
		a.rnd = acc.Rnd                      //update the highest seen round
		for k, value := range a.historyVal { //here value is PromiseSlot
			if acc.Val == value.Vval && acc.Slot > value.ID {
				delete(a.historyVal, k)
			}
		}
		a.historyVal[acc.Slot] = PromiseSlot{ID: acc.Slot,
			Vrnd: acc.Rnd,
			Vval: acc.Val}

		lrn.From = a.id
		lrn.Slot = acc.Slot
		lrn.Rnd = acc.Rnd
		lrn.Val = acc.Val
		output = true
	} else {
		output = false
	}
	return lrn, output
}

