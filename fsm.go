package fsm

import (
	"errors"
)

// StateNode represents a state node in FSM graph
type StateNode int

// EventType represents an event that makes the state transfer from one to another
type EventType int

// Event is the concrete structure
type Event interface {
	// Type returns the event type
	Type() EventType
	// Unwrap returns the original event object
	Unwrap() interface{}
}

// TransEdge describes the tranformation relationship between two states
// state From recieves an Event and then transfers to state To
// If Callback is setted, it will be invoked each time the transformation occurs,
// and only when Callback returns true, will the transformation happen.
type TransEdge struct {
	From     StateNode
	Event    EventType
	To       StateNode
	Callback Callback
}

// Callback is the signature of callback function that will be invoked when
// state transformation occurs
type Callback func(Event) bool

// StateMachine ...
type StateMachine interface {
	// Emit emits an Event and makes the transformation occur, thread unsafe
	// If state transfer successfully, return true, or return false
	Emit(Event) bool
	// CurrentState returns current state
	CurrentState() StateNode
}

// NewFSM is the constructor for a finite state machine
// our fsm will start at `st` state. If ignore isn't nil,
// it will be invoked when emit an unacceptable Event for current state
func NewFSM(st StateNode, edges []TransEdge, ignore Callback) (StateMachine, error) {
	m := &stateMachineImpl{
		graph:   make(map[StateNode]map[EventType]stateInfo),
		current: st,
		ignore:  ignore,
	}
	for _, edge := range edges {
		transTable, ok := m.graph[edge.From]
		if !ok {
			transTable = make(map[EventType]stateInfo)
			m.graph[edge.From] = transTable
		}
		if _, ok := transTable[edge.Event]; ok {
			return nil, errors.New("invalid fsm")
		}
		transTable[edge.Event] = stateInfo{
			state: edge.To,
			cb:    edge.Callback,
		}
	}
	return m, nil
}

type stateInfo struct {
	state StateNode
	cb    Callback
}

type stateMachineImpl struct {
	graph   map[StateNode]map[EventType]stateInfo
	current StateNode
	ignore  Callback
}

func (m *stateMachineImpl) Emit(e Event) bool {
	t := e.Type()
	nextState, ok := m.graph[m.current][t]
	if !ok {
		if m.ignore != nil {
			m.ignore(e)
		}
		return false
	}
	state, allowTrans := nextState.state, nextState.cb
	if allowTrans != nil {
		if allowTrans(e) {
			m.current = state
		} else {
			return false
		}
	} else {
		m.current = state
	}
	return true
}

func (m *stateMachineImpl) CurrentState() StateNode {
	return m.current
}
