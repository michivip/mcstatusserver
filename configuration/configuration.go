package configuration

import (
	"os"
	"encoding/json"
	"log"
)

func LoadConfiguration(fileName string) (*ServerConfiguration) {
	config := &ServerConfiguration{}
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating new configuration file: \"%v\"\n", fileName)
			file, err = os.Create(fileName)
			if err != nil {
				panic(err)
			}
			jsonEncoder := json.NewEncoder(file)
			jsonEncoder.SetIndent("", "  ")
			if err = jsonEncoder.Encode(getDefaultConfiguration()); err != nil {
				panic(err)
			}
			if err = file.Close(); err != nil {
				panic(err)
			}
			file, err = os.Open(fileName)
			if err != nil {
				panic(err)
			}
			log.Println("New config file created! Please adjust your values and restart the application.")
			os.Exit(0)
		} else {
			panic(err)
		}
	}
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		panic(err)
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	return config
}

func getDefaultConfiguration() *ServerConfiguration {
	return &ServerConfiguration{
		Address:           "localhost:25565",
		ConnectionTimeout: 10000,
		StatsHttpServer: StatsHttpServer{
			Addr:              "localhost:8080",
			StatisticsMapSize: 5,
			ErrorTimeout:      1000,
			Timezone:          "CET",
		},
		LogFile: "access.log",
		LoginAttempt: LoginAttemptValues{
			DisconnectText: ChatValue{
				Text:       "You are not ",
				Bold:       "true",
				Color:      "red",
				Insertion:  "true",
				Underlined: "true",
				Obfuscated: "true",
				Italic:     "true",
				Extra: []ChatComponentValue{
					{
						Text:  "allowed to access ",
						Color: "green",
					},
					{
						Text:   "this server.",
						Italic: "true",
					},
				},
			},
		},
		Motd: MessageOfTheDayValues{
			Version: struct {
				Name     string `json:"name"`
				Protocol int    `json:"protocol"`
			}{Name: "mcstatusserver 420", Protocol: -1},
			Players: struct {
				Max    int `json:"max"`
				Online int `json:"online"`
				Sample []struct {
					Name string `json:"name"`
					Id   string `json:"id"`
				} `json:"sample,omitempty"`
			}{Max: 1337, Online: 42, Sample: []struct {
				Name string `json:"name"`
				Id   string `json:"id"`
			}{{Name: "Hi there, this is", Id: "7cd21442-4bd9-4b02-8539-1a2c4771ed3c"}, {Name: "my public server", Id: "c87a3f54-3802-4227-b72a-17ee3f10cbd9"}}},
			Description: struct{ Text string `json:"text"` }{Text: "§cThis server runs with §aGo§c.\n§7https://github.com/michivip/mcstatusserver"},
			FaviconPath: "icon.png",
		},
	}
}
