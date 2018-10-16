# ggr-legacy-exporter
Prometheus exporter for Windows/Mac standalone nodes

```
go run main.go -surl http://192.168.1.1:4444/wd/hub/sessions
```

```
$ curl http://localhost:9156/metrics 
# HELP sessions_total number of sessions with all types browsers.
# TYPE sessions_total gauge
sessions_total{browsername="internet explorer",platform="WINDOWS",version="9"} 1
```

TODO: Support selenoid structure
