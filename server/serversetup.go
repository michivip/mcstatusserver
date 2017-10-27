package server

import (
	"net"
	"os"
	"github.com/michivip/mcstatusserver/datatypes"
	"log"
	"io"
	"bytes"
	"encoding/json"
	"github.com/satori/go.uuid"
)

const address = "localhost:25565"

func StartServer() *net.TCPListener {
	log.Printf("Starting server on %v\n", address)
	tcpAddress, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp4", tcpAddress)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	return listener
}

func WaitForConnections(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn *net.TCPConn) {
	// initial state is Handshaking (http://wiki.vg/Protocol#Definitions)
	States[conn] = HandshakingState
	// infinite loop of packet reading
	for {
		if packet, err, _ := datatypes.ReadPacket(conn); err != nil {
			if err == io.EOF {
				log.Printf("Closed connection to %v.", conn.RemoteAddr())
				break
			} else {
				panic(err)
			}
		} else {
			log.Printf("Received packet with id %v from %v.", packet.Id, conn.RemoteAddr())
			switch packet.Id {
			case 0:
				handleHandshakePacket(conn, packet)
				break
			case 1:
				handlePingPacket(conn, packet)
				break
			default:
				log.Printf("Received packet with unknown ID: %v\n", packet.Id)
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

func handleHandshakePacket(conn *net.TCPConn, packet datatypes.Packet) {
	currentState := States[conn]
	switch currentState {
	case HandshakingState:
		version, err, _ := datatypes.ReadVarInt(packet.Content)
		if err != nil {
			log.Printf("Received invalid version data from %v: %v", conn.RemoteAddr(), err)
			return
		}
		address, err, _ := datatypes.ReadString(packet.Content)
		if err != nil {
			log.Printf("Received invalid address data from %v: %v", conn.RemoteAddr(), err)
			return
		}
		port, err := datatypes.ReadUnsignedShort(packet.Content)
		if err != nil {
			log.Printf("Received invalid port data from %v: %v", conn.RemoteAddr(), err)
			return
		}
		nextRawState, err, _ := datatypes.ReadVarInt(packet.Content)
		var nextState ConnectionState
		if err != nil {
			log.Printf("Received invalid next state data from %v: %v", conn.RemoteAddr(), err)
			return
		} else if nextState, err = GetConnectionStateFromInt(nextRawState); err != nil {
			log.Printf("The sent next state from %v is invalid: %v", conn.RemoteAddr(), err)
			return
		} else {
			States[conn] = nextState
		}
		log.Printf("Received handshake packet. [version=%v, address=%v, port=%v, nextRawState=%v]\n", version, address, port, nextRawState)
		return
	case StatusState:
		// no additional data is sent which can be read
		data, err := json.Marshal(&StatusResponse{
			Version: struct {
				Name     string `json:"name"`
				Protocol int    `json:"protocol"`
			}{Name: "mcstatusserver", Protocol: -1},
			Players: struct {
				Max    int `json:"max"`
				Online int `json:"online"`
				Sample []struct {
					Name string `json:"name"`
					Id   string `json:"id"`
				} `json:"sample,omitempty"`
			}{Max: 1337, Online: 0, Sample: []struct {
				Name string `json:"name"`
				Id   string `json:"id"`
			}{{Name: "Hi there, this is", Id: uuid.NewV4().String()}, {Name: "my public server", Id: uuid.NewV4().String()}}},
			Description: struct{ Text string `json:"text"` }{Text: "§cThis server runs with §aGo§c.\n§7https://github.com/michivip/mcstatusserver"},
		})
		if err != nil {
			panic(err)
		} else {
			buffer := bytes.NewBuffer([]byte{})
			if err, _ := datatypes.WriteString(buffer, string(data)); err != nil {
				panic(err)
			}
			err, _ := datatypes.WritePacket(conn, datatypes.Packet{Content: buffer, Id: 0})
			if err != nil {
				panic(err)
			}
		}
		return
	default:
		log.Printf("Cannot handle received Handshake packet with current state: %v", currentState)
		return
	}
}

func handlePingPacket(conn *net.TCPConn, packet datatypes.Packet) {
	payload, err := datatypes.ReadLong(packet.Content)
	if err != nil {
		log.Printf("Received invalid ping payload data from %v: %v", conn.RemoteAddr(), err)
		return
	}
	payloadBuffer := bytes.NewBuffer([]byte{})
	if err = datatypes.WriteLong(payloadBuffer, payload); err != nil {
		panic(err)
	}
	if err, _ = datatypes.WritePacket(conn, datatypes.Packet{Content: payloadBuffer, Id: 1}); err != nil {
		panic(err)
	}
}
