package server

import (
	"net"
	"github.com/michivip/mcstatusserver/datatypes"
	"log"
	"io"
	"bytes"
	"encoding/json"
	"strings"
	"github.com/michivip/mcstatusserver/configuration"
	"fmt"
	"time"
	"github.com/michivip/mcstatusserver/statsserver"
)

var Closed = false

func StartServer(config *configuration.ServerConfiguration) *net.TCPListener {
	log.Printf("Starting server on %v\n", config.Address)
	tcpAddress, err := net.ResolveTCPAddr("tcp4", config.Address)
	if err != nil {
		log.Fatalf("An error ocurred while resolving the bind address:")
		panic(err)
	}
	listener, err := net.ListenTCP("tcp4", tcpAddress)
	if err != nil {
		log.Fatalf("Could not listen to bind address %v:\n", config.Address)
		panic(err)
	}
	return listener
}

func WaitForConnections(listener *net.TCPListener, config *configuration.ServerConfiguration) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			if Closed {
				return
			} else {
				log.Fatalln("There was an error while accepting a TCP connection:")
				panic(err)
			}
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn *net.TCPConn, config *configuration.ServerConfiguration) {
	log.Printf("[%v] --> Incoming connection.", conn.RemoteAddr())
	var connectionOpen bool = true
	go func() {
		time.Sleep(time.Millisecond * time.Duration(config.ConnectionTimeout))
		connectionOpen = false
		err := conn.Close()
		if err == nil {
			log.Printf("[%v] Idle timeout exceeded.", conn.RemoteAddr())
		}
	}()
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[%v] Recovered from handle packet method %T: %v", conn.RemoteAddr(), rec, rec)
		}
		conn.Close()
		log.Printf("[%v] <-- Closed connection.", conn.RemoteAddr())
	}()
	// initial state is Handshaking (http://wiki.vg/Protocol#Definitions)
	States[conn] = HandshakingState
	// infinite loop of packet reading
	for {
		if packet, err, _ := datatypes.ReadPacket(conn); err != nil {
			if err == io.EOF {
				return
			} else if err == io.ErrUnexpectedEOF {
				log.Printf("[%v] Received invalid packet data.\n", conn.RemoteAddr())
				return
			} else if !connectionOpen {
				break
			} else {
				log.Printf("[%v] Unknown error while reading packet:\n", conn.RemoteAddr())
				panic(err)
			}
		} else {
			var packetHandleError ConnectionError
			switch packet.Id {
			case 0:
				packetHandleError = handleHandshakePacket(conn, packet, config)
				break
			case 1:
				packetHandleError = handlePingPacket(conn, packet)
				break
			default:
				log.Printf("[%v] Received packet with unknown ID: %v\n", conn.RemoteAddr(), packet.Id)
				return
			}
			if packetHandleError != nil {
				if packetHandleError.IsFatal() {
					log.Printf("[%v] A fatal error ocurred while handling a packet with the id %v:\n", conn.RemoteAddr(), packet.Id)
					panic(err)
				} else {
					log.Printf("[%v] Packet handle error ocurred: %T: %v\n", conn.RemoteAddr(), err, err.Error())
					return
				}
			}
		}
	}
}

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"sample,omitempty"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Favicon string `json:"favicon,omitempty"`
}

type ErrBasedConnectionError struct {
	ThrownErr error
	Fatal     bool
}

func (errBasedConnectionError ErrBasedConnectionError) Error() string {
	return errBasedConnectionError.ThrownErr.Error()
}

func (errBasedConnectionError ErrBasedConnectionError) IsFatal() bool {
	return errBasedConnectionError.Fatal
}

type ConnectionError interface {
	error
	IsFatal() bool
}

type ErrInvalidDataReceived struct {
	DataName string
}

func (errInvalidDataReceived ErrInvalidDataReceived) Error() string {
	return fmt.Sprintf("received invalid %v", errInvalidDataReceived.DataName)
}

func (errInvalidDataReceived ErrInvalidDataReceived) IsFatal() bool {
	return false
}

func handleHandshakePacket(conn *net.TCPConn, packet datatypes.Packet, config *configuration.ServerConfiguration) ConnectionError {
	currentState := States[conn]
	switch currentState {
	case HandshakingState:
		version, err, _ := datatypes.ReadVarInt(packet.Content)
		if err != nil {
			return ErrInvalidDataReceived{"protocol version"}
		}
		connectAddress, err, _ := datatypes.ReadString(packet.Content)
		if err != nil {
			return ErrInvalidDataReceived{"server connect address"}
		}
		port, err := datatypes.ReadUnsignedShort(packet.Content)
		if err != nil {
			return ErrInvalidDataReceived{"server port"}
		}
		nextRawState, err, _ := datatypes.ReadVarInt(packet.Content)
		var nextState ConnectionState
		if err != nil {
			return ErrInvalidDataReceived{"next state"}
		} else if nextState, err = GetConnectionStateFromInt(nextRawState); err != nil {
			return ErrBasedConnectionError{err, false}
		} else {
			States[conn] = nextState
		}
		log.Printf("[%v] Received handshake packet. [version=%v, connectAddress=%v, port=%v, nextRawState=%v]\n", conn.RemoteAddr(), version, connectAddress, port, nextRawState)
		return nil
	case StatusState:
		// no additional data is sent which can be read
		data, err := json.Marshal(&StatusResponse{
			Version:     config.Motd.Version,
			Players:     config.Motd.Players,
			Description: config.Motd.Description,
			Favicon:     config.Motd.FaviconPath,
		})
		if err != nil {
			return ErrBasedConnectionError{fmt.Errorf("could not serialize Handshake MOTD data: %v", err), true}
		} else {
			buffer := bytes.NewBuffer([]byte{})
			if err, _ := datatypes.WriteString(buffer, string(data)); err != nil {
				return ErrBasedConnectionError{err, false}
			}
			err, _ := datatypes.WritePacket(conn, datatypes.Packet{Content: buffer, Id: 0})
			if err != nil {
				return ErrBasedConnectionError{err, false}
			}
			statsserver.RegisterPing()
		}
		return nil
	case LoginState:
		playerName, err, _ := datatypes.ReadString(packet.Content)
		if err != nil {
			return ErrInvalidDataReceived{"player name"}
		}
		// maximum length of a player name is 16 ascii characters
		if len(playerName) > 16 {
			return ErrInvalidDataReceived{fmt.Sprintf("player name length %v", strings.Replace(playerName, "\\", "\\\\", -1))}
		}
		data := bytes.NewBuffer([]byte{})
		jsonBytes, err := json.Marshal(config.LoginAttempt.DisconnectText)
		if err, _ = datatypes.WriteString(data, string(jsonBytes)); err != nil {
			return ErrBasedConnectionError{err, false}
		}
		if err, _ = datatypes.WritePacket(conn, datatypes.Packet{Content: data, Id: 0}); err != nil {
			return ErrBasedConnectionError{err, false}
		}
		statsserver.RegisterLogin()
		return nil
	default:
		return ErrBasedConnectionError{fmt.Errorf("can not handle Handshake packet with current state: %v", currentState), false}
	}
}

func handlePingPacket(conn *net.TCPConn, packet datatypes.Packet) ConnectionError {
	payload, err := datatypes.ReadLong(packet.Content)
	if err != nil {
		return ErrInvalidDataReceived{"ping payload"}
	}
	payloadBuffer := bytes.NewBuffer([]byte{})
	if err = datatypes.WriteLong(payloadBuffer, payload); err != nil {
		return ErrBasedConnectionError{err, false}
	}
	if err, _ = datatypes.WritePacket(conn, datatypes.Packet{Content: payloadBuffer, Id: 1}); err != nil {
		return ErrBasedConnectionError{err, false}
	}
	return nil
}
