package web

import (
	"encoding/json"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/gorilla/mux"
	"net/http"
)

func registerApiRoutes(subrouter *mux.Router, getBridgeOrNil func() plugins.ICoreBridge, getState func() *State) {
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
		r.Context()
		err := bridge.StopSeeding(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodPost)

	// TODO: generate the config json object with subrouter.Walk
}
