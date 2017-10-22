package server

import (
	"net"
	"os"
	"fmt"
)

const address = "localhost:25565"

func StartServer() *net.TCPListener {
	fmt.Printf("Starting server on %v\n", address)
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
	buffer := make([]byte, 1024)
	for {
		conn.SetReadBuffer(1024)
		reqLen, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("An error occurred while handling incoming data: %v", err)
			break
		}
		if reqLen == 0 {
			break
		}
		fmt.Printf("ReqLen: %v\n", reqLen)
		fmt.Printf("Message: %v\n", string(buffer))
	}
}
