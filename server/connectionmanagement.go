package server

import (
	"net"
	"fmt"
	"sync"
)

var States map[*net.TCPConn]ConnectionState = make(map[*net.TCPConn]ConnectionState)
var StatesLock = sync.RWMutex{}

type ConnectionState uint8

const HandshakingState ConnectionState = ConnectionState(0)
const StatusState ConnectionState = ConnectionState(1)
const LoginState ConnectionState = ConnectionState(2)

type Connection struct {
	CurrentState ConnectionState
}

type ErrNoStateFound struct {
	RawState int
}

func (errNoStateFound ErrNoStateFound) Error() string {
	return fmt.Sprintf("could not find any state for value: %v", errNoStateFound.RawState)
}

func GetConnectionStateFromInt(rawState int) (ConnectionState, error) {
	switch rawState {
	case 0:
		return HandshakingState, nil
	case 1:
		return StatusState, nil
	case 2:
		return LoginState, nil
	default:
		return ConnectionState(0), ErrNoStateFound{RawState: rawState}
	}
}
