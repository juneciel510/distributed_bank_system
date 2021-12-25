package leaderdetector

// A MonLeaderDetector represents a Monarchical Eventual Leader Detector as
// described at page 53 in:
// Christian Cachin, Rachid Guerraoui, and Luís Rodrigues: "Introduction to
// Reliable and Secure Distributed Programming" Springer, 2nd edition, 2011.
type MonLeaderDetector struct { // TODO(student): Add needed fields

	nodeIDs []int
	susp    map[int]bool
	leader  int
	subs    []chan int
}

// NewMonLeaderDetector returns a new Monarchical Eventual Leader Detector
// given a list of node ids.
func NewMonLeaderDetector(nodeIDs []int) *MonLeaderDetector {
	// TODO(student): Add needed implementation
	m := &MonLeaderDetector{}
	m.nodeIDs = nodeIDs
	m.leader = UnknownID
	sus := make(map[int]bool)
	m.susp = sus
	m.subs = make([]chan int, 0)

	return m

}

// Leader returns the current leader. Leader will return UnknownID if all nodes
// are suspected.
func (m *MonLeaderDetector) Leader() int {
	// TODO(student): Implement
	var lead int = m.leader
	lead = m.max(m.nodeIDs)
	if lead != m.leader {
		for sub := range m.subs {
			m.subs[sub] <- lead
		}
	}
	m.leader = lead
	return m.leader

}

// Suspect instructs the leader detector to consider the node with matching
// id as suspected. If the suspect indication result in a leader change
// the leader detector should publish this change to its subscribers.
func (m *MonLeaderDetector) Suspect(id int) {
	// TODO(student): Implement
	m.susp[id] = true
	m.leader = m.Leader()
}

// Restore instructs the leader detector to consider the node with matching
// id as restored. If the restore indication result in a leader change
// the leader detector should publish this change to its subscribers.
func (m *MonLeaderDetector) Restore(id int) {
	// TODO(student): Implement
	m.susp[id] = false
	m.leader = m.Leader()

}

// Subscribe returns a buffered channel which will be used by the leader
// detector to publish the id of the highest ranking node.
// The leader detector will publish UnknownID if all nodes become suspected.
// Subscribe will drop publications to slow subscribers.
// Note: Subscribe returns a unique channel to every subscriber;
// it is not meant to be shared.
func (m *MonLeaderDetector) Subscribe() <-chan int {
	// TODO(student): Implement
	ch := make(chan int, len(m.nodeIDs))
	m.subs = append(m.subs, ch)
	return ch
}

// TODO(student): Add other unexported functions or methods if needed.

func (m *MonLeaderDetector) max(nodeid []int) int {
	var l int = UnknownID
	for _, val := range nodeid {
		if val > l && m.susp[val] == false {
			l = val

		}
	}
	return l
}

func (m *MonLeaderDetector) Reconfig(sIDL []int) {
	m.nodeIDs = sIDL
	m.subs = make([]chan int, 0)
}
