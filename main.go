package main

import (
	"fmt"
	"github.com/michivip/mcstatusserver/server"
)

const asciiArt = "                           _             _                                                           \n" +
	"  _ __ ___     ___   ___  | |_    __ _  | |_   _   _   ___   ___    ___   _ __  __   __   ___   _ __ \n" +
	" | '_ ` _ \\   / __| / __| | __|  / _` | | __| | | | | / __| / __|  / _ \\ | '__| \\ \\ / /  / _ \\ | '__|\n" +
	" | | | | | | | (__  \\__ \\ | |_  | (_| | | |_  | |_| | \\__ \\ \\__ \\ |  __/ | |     \\ V /  |  __/ | |   \n" +
	" |_| |_| |_|  \\___| |___/  \\__|  \\__,_|  \\__|  \\__,_| |___/ |___/  \\___| |_|      \\_/    \\___| |_|   \n" +
	"                                                                                                     "

func main() {
	fmt.Println(asciiArt)
	listener := server.StartServer()
	defer listener.Close()
	server.WaitForConnections(listener)
}
