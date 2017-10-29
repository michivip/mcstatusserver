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
	"strings"
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
			Favicon:     "data:image/png;base64,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",
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
	case LoginState:
		playerName, err, _ := datatypes.ReadString(packet.Content)
		if err != nil {
			log.Printf("Received invalid player name from %v: %v", conn.RemoteAddr(), err)
			return
		}
		// maximum length of a player name is 16 ascii characters
		if len(playerName) > 16 {
			log.Printf("Received invalid player name from %v: %v", conn.RemoteAddr(), strings.Replace(playerName, "\\", "\\\\", -1))
			return
		}
		data := bytes.NewBuffer([]byte{})
		if err, _ = datatypes.WriteString(data, "{\"text\":\"You can not join the server at the moment.\"}"); err != nil {
			panic(err)
		}
		if err, _ = datatypes.WritePacket(conn, datatypes.Packet{Content: data, Id: 0}); err != nil {
			panic(err)
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
