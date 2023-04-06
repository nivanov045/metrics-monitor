package api

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type api struct {
	service Service
}

func New(service Service) *api {
	return &api{service: service}
}

var _ API = &api{}

func (a *api) Run(address string) error {
	log.Info().Interface("address", address).Msg("server started")

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
	log.Debug().Msg("updating of metrics started")

	w.Header().Set("content-type", "application/json")

	defer r.Body.Close()
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Stack()

		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = a.service.ParseAndSave(respBody)
	if err != nil {
		log.Error().Err(err)

		switch err.Error() {
		case "wrong metrics type":
			w.WriteHeader(http.StatusNotImplemented)
		case "can't parse value", "wrong hash":
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write([]byte("{}"))
		return
	}

	log.Debug().Msg("parsed and saved")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (a *api) getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("getting of metrics started")

	w.Header().Set("content-type", "application/json")

	defer r.Body.Close()
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Stack()
		w.WriteHeader(http.StatusNotFound)
		return
	}

	val, err := a.service.ParseAndGet(respBody)
	if err != nil {
		log.Error().Err(err)

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

	log.Debug().Msg("parsed and get")

	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func (a *api) rootHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msg("rootHandler started")

	w.Header().Set("content-type", "text/html")
	for _, val := range a.service.GetKnownMetrics() {
		w.Write([]byte(val + "\n"))
	}
}

func (a *api) pingDBHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msg("pingDBHandler started")

	w.Header().Set("content-type", "text/html")

	if !a.service.IsDBConnected() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *api) updatesMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("updatesMetricsHandler started")

	w.Header().Set("content-type", "application/json")

	defer r.Body.Close()
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Stack()

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}

	err = a.service.ParseAndSaveSeveral(respBody)
	if err != nil {
		log.Error().Err(err).Stack()

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}
	log.Debug().Msg("parsed and saved several")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
