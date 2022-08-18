package api

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	ParseAndSave([]byte) error
	ParseAndGet([]byte) ([]byte, error)
	GetKnownMetrics() []string
	IsDBConnected() bool
	ParseAndSaveSeveral([]byte) error
}

type api struct {
	service Service
}

func New(service Service) *api {
	return &api{service: service}
}

func (a *api) Run(address string) error {
	log.Println("api::Run::info: started with addr:", address)
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json", "text/html"))

	r.Post("/update/", a.updateMetricsHandler)
	r.Get("/", a.rootHandler)
	r.Post("/value/", a.getMetricsHandler)
	r.Get("/ping", a.pingDBHandler)
	r.Post("/updates/", a.updatesMetricsHandler)
	return http.ListenAndServe(address, r)
}

func (a *api) updateMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::updateMetricsHandler::info: started")
	w.Header().Set("content-type", "application/json")
	defer r.Body.Close()
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("api::updateMetricsHandler::warning: can't read response body with:", err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}
	if err := a.service.ParseAndSave(respBody); err == nil {
		log.Println("api::updateMetricsHandler::info: parsed and saved")
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("api::updateMetricsHandler::warning: in parsing and saving:", err)
		if err.Error() == "wrong metrics type" {
			w.WriteHeader(http.StatusNotImplemented)
		} else if err.Error() == "can't parse value" || err.Error() == "wrong hash" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
	w.Write([]byte("{}"))
}

func (a *api) getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::getMetricsHandler::info: started")
	w.Header().Set("content-type", "application/json")
	defer r.Body.Close()
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("api::getMetricsHandler::warning: can't read response body with:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if val, err := a.service.ParseAndGet(respBody); err == nil {
		log.Println("api::getMetricsHandler::info: parsed and get")
		w.WriteHeader(http.StatusOK)
		w.Write(val)
	} else {
		log.Println("api::getMetricsHandler::warning: in parsing:", err)
		if err.Error() == "wrong metrics type" {
			w.WriteHeader(http.StatusNotImplemented)
		} else if err.Error() == "no such metric" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte("{}"))
	}
}

func (a *api) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::rootHandler::info: started")
	w.Header().Set("content-type", "text/html")
	for _, val := range a.service.GetKnownMetrics() {
		w.Write([]byte(val + "\n"))
	}
}

func (a *api) pingDBHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::pingDBHandler::info: started")
	w.Header().Set("content-type", "text/html")
	if a.service.IsDBConnected() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *api) updatesMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::updatesMetricsHandler::info: started")
	w.Header().Set("content-type", "application/json")
	defer r.Body.Close()
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("api::updatesMetricsHandler::warning can't read response body with:", err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}
	if err := a.service.ParseAndSaveSeveral(respBody); err == nil {
		log.Println("api::updatesMetricsHandler::info: parsed and saved")
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("api::updatesMetricsHandler::warning: wrong group query:", err)
		w.WriteHeader(http.StatusNotFound)
	}
	w.Write([]byte("{}"))
}

type API interface {
	Run(string2 string) error
}

var _ API = &api{}
