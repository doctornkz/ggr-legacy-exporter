package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Sessions - main struct of Selenium hub response
type Sessions struct {
	Class     string
	Status    int
	State     string
	SessionID string
	HCode     int64
	ID        string
	Value     []struct {
		Capabilities struct {
			Version     string
			Platform    string
			BrowserName string
		}
	}
}

var paths = struct {
	Metrics string
}{
	Metrics: "/metrics",
}

var (
	sessions    Sessions
	seleniumURL string
	listen      string
)

func init() {
	flag.StringVar(&seleniumURL, "surl", "http://localhost:4444/wd/hub/sessions", "Full URL to session handler")
	flag.StringVar(&listen, "listen", "localhost:9156", "Host and port to listen to")
	flag.Parse()
}

func main() {
	http.HandleFunc(paths.Metrics, func(w http.ResponseWriter, r *http.Request) {
		sessions, err := sessionsReader()
		if err != nil {
			w.Write([]byte("Problem with selenium's response, see logs for details."))
		} else {
			metricPage := formatter(sessions)
			w.Write([]byte(metricPage))
		}
	})
	if err := http.ListenAndServe(listen, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}

func formatter(sessions Sessions) string {

	var outputString string
	// sessions_windows_internetexplorer_10 = N
	browsers := make(map[string]int)
	var browserMetric string
	// TODO: Too much complicated
	for _, v := range sessions.Value {
		platform := strings.ToLower(v.Capabilities.Platform)
		browser := strings.Replace(strings.ToLower(v.Capabilities.BrowserName), " ", "", -1)
		version := v.Capabilities.Version
		browserMetric = fmt.Sprintf("sessions_%s_%s_%s", platform, browser, version)
		browsers[browserMetric]++
	}

	for k, v := range browsers {
		outputString = outputString + fmt.Sprintf("%s %d\n", k, v)
	}

	// Common metrics
	outputString = outputString + fmt.Sprintf("sessions_state %s\nsessions_status %d\n", sessions.State, sessions.Status)
	return outputString
}

func sessionsReader() (Sessions, error) {
	resp, err := http.Get(seleniumURL) // TODO: remake errors
	if err != nil {
		log.Printf("Problem with HTTP request : %v", err)
		return sessions, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Can't read response from socket: %v", err)
		return sessions, err

	}
	err = json.Unmarshal(body, &sessions)

	if err != nil {
		log.Printf("Can't Unmarshal structure, check URL : %v", err)
		return sessions, err

	}
	return sessions, nil
}
