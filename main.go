package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SeleniumSessions - main struct of Selenium hub response
type SeleniumSessions struct {
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

// SessionLabeled - data structure for export metrics in Prometheus runtime
type SessionLabeled struct {
	BrowserName string
	Platform    string
	Version     string
}

var paths = struct {
	Metrics string
}{
	Metrics: "/metrics",
}

var (
	seleniumSessions SeleniumSessions
	seleniumURL      string
	listen           string
	version          bool
)

var (
	sessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sessions_total",
			Help: "number of sessions with all types browsers.",
		}, []string{"platform", "browsername", "version"})
)

const (
	currentVersion     = "0.1.0"
	defaultSeleniumURL = "http://localhost:8097/wd/hub/sessions" // By default 4444 is using
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

	prometheus.MustRegister(sessions)

}

func main() {

	go func() {
		for {
			currentSession := getSessions()
			for k, v := range currentSession {
				sessions.WithLabelValues(k.Platform, k.BrowserName, k.Version).Set(float64(v))
			}

			time.Sleep(time.Duration(time.Second * 10))
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listen, nil))
}

func getSessions() map[SessionLabeled]int {
	seleniumSessions, err := sessionsReader()
	if err != nil {
		log.Printf("Problem with Session generate : %v", err)
		return nil
	}

	sessionsLabeled := make(map[SessionLabeled]int)
	for _, v := range seleniumSessions.Value {
		sessionsLabeled[SessionLabeled{BrowserName: v.Capabilities.BrowserName,
			Platform: v.Capabilities.Platform,
			Version:  v.Capabilities.Version}]++
	}
	return sessionsLabeled
}

func sessionsReader() (SeleniumSessions, error) {
	resp, err := http.Get(seleniumURL) // TODO: remake errors
	if err != nil {
		log.Printf("Problem with HTTP request : %v", err)
		return seleniumSessions, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Can't read response from socket: %v", err)
		return seleniumSessions, err

	}
	err = json.Unmarshal(body, &seleniumSessions)

	if err != nil {
		log.Printf("Can't Unmarshal structure, check URL : %v", err)
		return seleniumSessions, err

	}
	return seleniumSessions, nil
}
