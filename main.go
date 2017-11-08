package main

import (
	"github.com/michivip/mcstatusserver/server"
	"flag"
	"github.com/michivip/mcstatusserver/configuration"
	"log"
	"os"
	"io/ioutil"
	"encoding/base64"
	"io"
	"bufio"
	"github.com/michivip/mcstatusserver/statsserver"
	"time"
)

const asciiArt = "                           _             _                                                           \n" +
	"  _ __ ___     ___   ___  | |_    __ _  | |_   _   _   ___   ___    ___   _ __  __   __   ___   _ __ \n" +
	" | '_ ` _ \\   / __| / __| | __|  / _` | | __| | | | | / __| / __|  / _ \\ | '__| \\ \\ / /  / _ \\ | '__|\n" +
	" | | | | | | | (__  \\__ \\ | |_  | (_| | | |_  | |_| | \\__ \\ \\__ \\ |  __/ | |     \\ V /  |  __/ | |   \n" +
	" |_| |_| |_|  \\___| |___/  \\__|  \\__,_|  \\__|  \\__,_| |___/ |___/  \\___| |_|      \\_/    \\___| |_|   \n" +
	"                                                                                                     \n"

func main() {
	configurationFile := flag.String("config", "config.json", "The path to your custom configuration logFile.")
	flag.Parse()

	os.Stdout.WriteString(asciiArt)
	config := configuration.LoadConfiguration(*configurationFile)
	encodedFavicon := loadFavicon(config)
	config.Motd.FaviconPath = encodedFavicon
startLog:
	logFile, err := os.OpenFile(config.LogFile, os.O_RDWR, os.ModePerm)
	if os.IsNotExist(err) {
		logFile, err = os.Create(config.LogFile)
		if err != nil {
			log.Fatalf("There was an error while creating the logging file (%v):\n", config.LogFile)
			panic(err)
		}
	} else if err != nil {
		log.Fatalf("There was an error while opening the logging file (%v):\n", config.LogFile)
		panic(err)
	} else {
		logFile.Close()
		if err = os.Remove(config.LogFile); err != nil {
			log.Fatalf("There was an error while removing the old logging file (%v):\n", config.LogFile)
			panic(err)
		}
		goto startLog
	}
	log.SetOutput(ConsoleFileWriter{logFile})
	listener := server.StartServer(config)
	statsServer, err := statsserver.SetupServer(config)
	defer func() {
		log.Println("Shutting down server...")
		listener.Close()
		log.Println("Shutting down stats http server...")
		statsServer.Close()
		log.Println("Closing log file...")
		logFile.Close()
	}()
	go server.WaitForConnections(listener, config)
	errorChannel := make(chan error)
	go func() {
		err := statsServer.ListenAndServe()
		if err != nil && errorChannel != nil {
			errorChannel <- err
		}
	}()
	select {
	case err := <-errorChannel:
		log.Println("There was an error while running the stats server:")
		panic(err)
	case <-time.After(time.Millisecond * time.Duration(config.StatsHttpServer.ErrorTimeout)):
		close(errorChannel)
		errorChannel = nil
		break
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if text == "stop\n" || text == "close\n" {
			server.Closed = true
			break
		}
	}
}

type SyncWriter interface {
	io.Writer
	Sync() error
}

type ConsoleFileWriter struct {
	FileWriter SyncWriter
}

func (consoleFileWriter ConsoleFileWriter) Write(p []byte) (n int, err error) {
	n, err = consoleFileWriter.FileWriter.Write(p)
	if err != nil {
		panic(err)
		return
	}
	err = consoleFileWriter.FileWriter.Sync()
	if err != nil {
		panic(err)
		return
	}
	n, err = os.Stdout.Write(p)
	return
}

func loadFavicon(config *configuration.ServerConfiguration) string {
	faviconPath := config.Motd.FaviconPath
	if faviconPath == "" {
		return ""
	}
	faviconFile, err := os.Open(faviconPath)
	if err != nil {
		panic(err)
	}
	faviconBytes, err := ioutil.ReadAll(faviconFile)
	if err != nil {
		panic(err)
	}
	base64Favicon := base64.RawStdEncoding.EncodeToString(faviconBytes)
	if err := faviconFile.Close(); err != nil {
		panic(err)
	}
	return "data:image/png;base64," + base64Favicon
}
