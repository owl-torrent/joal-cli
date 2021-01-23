package web

import (
	"encoding/json"
	"github.com/anthonyraymond/joal-cli/internal/plugins/types"
	"github.com/gorilla/mux"
	"net/http"
)

func registerApiRoutes(subrouter *mux.Router, getBridgeOrNil func() types.ICoreBridge, getState func() *state) {
	subrouter.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(getState())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodGet)

	subrouter.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		err := bridge.StartSeeding()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodPost)

	subrouter.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		err := bridge.StopSeeding(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodPost)

	subrouter.HandleFunc("/configuration", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		conf, err := bridge.GetCoreConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(conf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodGet)

	subrouter.HandleFunc("/configuration", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		var userConf *types.RuntimeConfig
		err := json.NewDecoder(r.Body).Decode(userConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		conf, err := bridge.UpdateCoreConfig(userConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(conf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodPut)

	subrouter.HandleFunc("/torrent", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// TODO: implement
		w.WriteHeader(http.StatusTeapot)
	}).Methods(http.MethodPost)

	subrouter.HandleFunc("/torrent", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// TODO: implement
		w.WriteHeader(http.StatusTeapot)
	}).Methods(http.MethodDelete)

	subrouter.HandleFunc("/clients/all", func(w http.ResponseWriter, r *http.Request) {
		bridge := getBridgeOrNil()
		if bridge == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		clients, err := bridge.ListClientFiles()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(clients)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodGet)
}
