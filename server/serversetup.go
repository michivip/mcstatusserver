package server

import (
	"net"
	"os"
	"github.com/michivip/mcstatusserver/datatypes"
	"log"
	"io"
	"bytes"
	"encoding/json"
	"strings"
	"github.com/michivip/mcstatusserver/configuration"
)

func StartServer(config *configuration.ServerConfiguration) *net.TCPListener {
	log.Printf("Starting server on %v\n", config.Address)
	tcpAddress, err := net.ResolveTCPAddr("tcp4", config.Address)
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

func WaitForConnections(listener *net.TCPListener, config *configuration.ServerConfiguration) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn *net.TCPConn, config *configuration.ServerConfiguration) {
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
				handleHandshakePacket(conn, packet, config)
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

func handleHandshakePacket(conn *net.TCPConn, packet datatypes.Packet, config *configuration.ServerConfiguration) {
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
			Version:     config.Motd.Version,
			Players:     config.Motd.Players,
			Description: config.Motd.Description,
			Favicon:     config.Motd.FaviconPath,
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
		jsonBytes, err := json.Marshal(config.LoginAttempt.DisconnectText)
		if err, _ = datatypes.WriteString(data, string(jsonBytes)); err != nil {
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
