package main

import (
	"fmt"
	"github.com/michivip/mcstatusserver/server"
	"flag"
	"github.com/michivip/mcstatusserver/configuration"
	"log"
	"os"
	"io/ioutil"
	"encoding/base64"
)

const asciiArt = "                           _             _                                                           \n" +
	"  _ __ ___     ___   ___  | |_    __ _  | |_   _   _   ___   ___    ___   _ __  __   __   ___   _ __ \n" +
	" | '_ ` _ \\   / __| / __| | __|  / _` | | __| | | | | / __| / __|  / _ \\ | '__| \\ \\ / /  / _ \\ | '__|\n" +
	" | | | | | | | (__  \\__ \\ | |_  | (_| | | |_  | |_| | \\__ \\ \\__ \\ |  __/ | |     \\ V /  |  __/ | |   \n" +
	" |_| |_| |_|  \\___| |___/  \\__|  \\__,_|  \\__|  \\__,_| |___/ |___/  \\___| |_|      \\_/    \\___| |_|   \n" +
	"                                                                                                     "

func main() {
	configurationFile := flag.String("config", "config.json", "The path to your custom configuration file.")
	flag.Parse()

	fmt.Println(asciiArt)
	config, err := configuration.LoadConfiguration(*configurationFile)
	if err != nil {
		log.Printf("There was an error while loading the configuration: %v\n", err)
		os.Exit(1)
	}
	encodedFavicon, err := loadFavicon(config)
	if err != nil {
		log.Printf("There was an error while loading the favicon: %v\n", err)
		os.Exit(1)
	}
	config.Motd.FaviconPath = encodedFavicon
	listener := server.StartServer(config)
	defer listener.Close()
	server.WaitForConnections(listener, config)
}

func loadFavicon(config *configuration.ServerConfiguration) (string, error) {
	faviconPath := config.Motd.FaviconPath
	if faviconPath == "" {
		return faviconPath, nil
	}
	faviconFile, err := os.Open(faviconPath)
	if err != nil {
		return faviconPath, err
	}
	faviconBytes, err := ioutil.ReadAll(faviconFile)
	if err != nil {
		return faviconPath, err
	}
	base64Favicon := base64.RawStdEncoding.EncodeToString(faviconBytes)
	if err := faviconFile.Close(); err != nil {
		return faviconPath, err
	}
	return "data:image/png;base64," + base64Favicon, nil
}
