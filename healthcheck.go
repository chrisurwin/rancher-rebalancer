package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	router          = mux.NewRouter()
	healthcheckPort = ":9777"
)

func startHealthcheck() {
	router.HandleFunc("/ping", healthcheck).Methods("GET", "HEAD").Name("Healthcheck")
	logrus.Info("Healthcheck handler is listening on ", healthcheckPort)
	logrus.Fatal(http.ListenAndServe(healthcheckPort, router))
}

func healthcheck(w http.ResponseWriter, req *http.Request) {
	// 1) test controller
	w.Write([]byte("pong"))
}
