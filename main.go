package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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
	version     bool
)

const (
	currentVersion     = "0.0.4"
	defaultSeleniumURL = "http://localhost:4444/wd/hub/sessions"
	defaultListen      = "0.0.0.0:9156"
)

func init() {
	flag.StringVar(&seleniumURL, "surl", defaultSeleniumURL, "Full URL to session handler")
	flag.StringVar(&listen, "listen", defaultListen, "Host and port to listen to")
	flag.BoolVar(&version, "version", false, "Show version and exit")
	flag.Parse()
	if version {
		fmt.Printf("Selenium exporter v%s, see at github.com/doctornkz/ggr-legacy-exporter\n", currentVersion)
		os.Exit(0)
	}

	log.Printf("Selenium exporter v(%s) is running with parameters: selenium URL:%s, listening IP:PORT :%s\n",
		currentVersion, seleniumURL, listen)

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
	regexp.MustCompile("success")
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
	// SessionState couldn't be string, only float/int

	sessionsStateSuccess := 0
	if regexp.MustCompile("success").MatchString(sessions.State) {
		sessionsStateSuccess = 1
	}

	outputString = outputString + fmt.Sprintf("sessions_state %d\nsessions_status_success %d\n", sessions.Status, sessionsStateSuccess)
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
