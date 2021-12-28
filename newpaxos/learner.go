package multipaxos

// Learner represents a learner as defined by the Multi-Paxos algorithm.
type Learner struct { 

	id int // id: The id of the node running this instance of a Paxos acceptor.
	quorum int
	n      int

	slotLearn map[SlotID][]*Learn //to store the slotID and its learn

	outputTrue map[SlotID]bool //if the slot  have the true output, set the value of map to be true

	decidedOut chan<- DecidedValue
	learnIn    chan Learn

	stop chan struct{}
}

// NewLearner returns a new Multi-Paxos learner. It takes the
// following arguments:
//
// id: The id of the node running this instance of a Paxos learner.
//
// nrOfNodes: The total number of Paxos nodes.
//
// decidedOut: A send only channel used to send values that has been learned,
// i.e. decided by the Paxos nodes.
func NewLearner(id int, nrOfNodes int, decidedOut chan<- DecidedValue) *Learner {
	return &Learner{
		id: id,
		quorum:     (nrOfNodes / 2) + 1,
		n:          nrOfNodes,
		slotLearn:  make(map[SlotID][]*Learn),
		outputTrue: make(map[SlotID]bool),

		decidedOut: decidedOut,
		learnIn:    make(chan Learn, 8),

		stop: make(chan struct{}),
	}
}

// Start starts l's main run loop as a separate goroutine. The main run loop
// handles incoming learn messages.
func (l *Learner) Start() {
	go func() {
		for {
			select {
			case lrn := <-l.learnIn:
				val, sid, output := l.handleLearn(lrn)
				if !output {
					continue
				}
				l.decidedOut <- DecidedValue{SlotID: sid, Value: val}
			case <-l.stop:
				return
			}
		}
	}()
}

// Stop stops l's main run loop.
func (l *Learner) Stop() {
	l.stop <- struct{}{}
}

// DeliverLearn delivers learn lrn to learner l.
func (l *Learner) DeliverLearn(lrn Learn) {
	l.learnIn <- lrn
}

// Internal: handleLearn processes learn lrn according to the Multi-Paxos
// algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true, sid the id for the
// slot that was decided and val contain the decided value. If handleLearn
// returns false as output, then val and sid will have their zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, sid SlotID, output bool) {
	if l.slotLearn[learn.Slot] == nil {
		l.slotLearn[learn.Slot] = append(l.slotLearn[learn.Slot], &learn)
	} else {
		for _, v := range l.slotLearn[learn.Slot] {
			if &learn == v {
				output = false
				return val, sid, output
			} else if learn.From != v.From && learn.Slot == v.Slot && learn.Rnd == v.Rnd && learn.Val == v.Val {
				l.slotLearn[learn.Slot] = append(l.slotLearn[learn.Slot], &learn)
			} else if learn.Slot == v.Slot && learn.Rnd > v.Rnd {
				l.slotLearn[learn.Slot] = nil
				l.slotLearn[learn.Slot] = append(l.slotLearn[learn.Slot], &learn)

			}

		}

	}

	var lenLearn int

	lenLearn = len(l.slotLearn[learn.Slot])
	if lenLearn >= l.quorum {
		//to check if the slot already had the true output, if so, then output false
		if l.outputTrue[learn.Slot] == false {
			sid = learn.Slot
			val = l.slotLearn[learn.Slot][1].Val
			output = true
			l.outputTrue[sid] = true
		}

	} else {
		output = false
	}

	return val, sid, output

}

//exported method reconfig
func (l *Learner) Reconfig(nrOfNodes int) {
	l.quorum = (nrOfNodes / 2) + 1
	l.n = nrOfNodes
}
