package statsserver

import (
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"github.com/michivip/mcstatusserver/configuration"
	"encoding/json"
)

var statsSlice []*StatsData

const HourDateFormat = "2006-01-02-15"

type StatsData struct {
	HourDate     string `json:"hourDate"`
	PingCounter  int    `json:"pingCount"`
	LoginCounter int    `json:"loginCount"`
}

func SetupServer(config *configuration.ServerConfiguration) (*http.Server, error) {
	statsSlice = make([]*StatsData, config.StatsHttpServer.StatisticsMapSize)
	unixNano := time.Now().UnixNano()
	for i := 0; i < config.StatsHttpServer.StatisticsMapSize; i++ {
		unixNano -= int64(int64(time.Hour))
		statsSlice[config.StatsHttpServer.StatisticsMapSize-i-1] = &StatsData{time.Unix(0, unixNano).Format(HourDateFormat), 0, 0}
	}
	go func() {
		for {
			checkSlices()
			time.Sleep(time.Minute * 50)
		}
	}()
	statsSiteBytes, err := statsHtmlBytes()
	if err != nil {
		return &http.Server{}, err
	}
	router := mux.NewRouter()
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(200)
		writer.Write(statsSiteBytes)
	})
	router.HandleFunc("/stats/packets", func(writer http.ResponseWriter, request *http.Request) {
		checkSlices()
		writer.Header().Set("Content-Type", "text/json")
		writer.WriteHeader(http.StatusOK)
		data, err := json.Marshal(statsSlice)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			panic(err)
			return
		}
		writer.Write(data)
	})
	srv := &http.Server{
		Addr: config.StatsHttpServer.Addr,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Webserver", "mcstatusserver")
			router.ServeHTTP(writer, request)
		}),
	}
	return srv, nil
}

func checkSlices() {
	registerCount(0, 0)
}

func RegisterPing() {
	registerCount(1, 0)
}

func RegisterLogin() {
	registerCount(0, 1)
}

func registerCount(pingAmount, loginAmount int) {
	// Year-Month-Day-Hour
	hourDate := time.Now().Format(HourDateFormat)
	if len(statsSlice) > 0 {
		data := statsSlice[len(statsSlice)-1]
		if data.HourDate == hourDate {
			data.PingCounter += pingAmount
			data.LoginCounter += loginAmount
			return
		}
	}
	newStatsData := &StatsData{hourDate, pingAmount, loginAmount}
	copy(statsSlice, (statsSlice)[1:])
	statsSlice[len(statsSlice)-1] = newStatsData
}
