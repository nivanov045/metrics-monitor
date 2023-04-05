package api

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type api struct {
	service Service
}

func New(service Service) *api {
	return &api{service: service}
}

var _ API = &api{}

func (a *api) Run(address string) error {
	log.Println("api::Run::info: started with addr:", address)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json", "text/html"))

	r.Post("/update/", a.updateMetricsHandler)
	r.Post("/updates/", a.updatesMetricsHandler)
	r.Post("/value/", a.getMetricsHandler)

	r.Get("/", a.rootHandler)
	r.Get("/ping", a.pingDBHandler)

	return http.ListenAndServe(address, r)
}

func (a *api) updateMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::updateMetricsHandler::info: started")

	w.Header().Set("content-type", "application/json")
	w.Write([]byte("{}"))

	defer r.Body.Close()
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("api::updateMetricsHandler::warning: can't read response body with:", err)

		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = a.service.ParseAndSave(respBody)
	if err != nil {
		log.Println("api::updateMetricsHandler::warning: in parsing and saving:", err)

		switch err.Error() {
		case "wrong metrics type":
			w.WriteHeader(http.StatusNotImplemented)
		case "can't parse value", "wrong hash":
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}

	log.Println("api::updateMetricsHandler::info: parsed and saved")

	w.WriteHeader(http.StatusOK)
}

func (a *api) getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::getMetricsHandler::info: started")

	w.Header().Set("content-type", "application/json")

	defer r.Body.Close()
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("api::getMetricsHandler::warning: can't read response body with:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	val, err := a.service.ParseAndGet(respBody)
	if err != nil {
		log.Println("api::getMetricsHandler::warning: in parsing:", err)

		switch err.Error() {
		case "wrong metrics type":
			w.WriteHeader(http.StatusNotImplemented)
		case "no such metric":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}

		w.Write([]byte("{}"))
		return
	}

	log.Println("api::getMetricsHandler::info: parsed and get")

	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func (a *api) rootHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("api::rootHandler::info: started")

	w.Header().Set("content-type", "text/html")
	for _, val := range a.service.GetKnownMetrics() {
		w.Write([]byte(val + "\n"))
	}
}

func (a *api) pingDBHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("api::pingDBHandler::info: started")

	w.Header().Set("content-type", "text/html")

	if !a.service.IsDBConnected() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *api) updatesMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("api::updatesMetricsHandler::info: started")

	w.Header().Set("content-type", "application/json")

	defer r.Body.Close()
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("api::updatesMetricsHandler::warning can't read response body with:", err)

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}

	err = a.service.ParseAndSaveSeveral(respBody)
	if err != nil {
		log.Println("api::updatesMetricsHandler::warning: wrong group query:", err)

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}
	log.Println("api::updatesMetricsHandler::info: parsed and saved")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
