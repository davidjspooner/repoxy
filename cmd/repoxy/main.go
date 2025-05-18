package main

import (
    "net/http"
    "log"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    http.Handle("/metrics", promhttp.Handler())
    log.Println("Starting Repoxy on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
