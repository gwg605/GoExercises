package cron

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"valerygordeev/go/exercises/libs/base"

	"github.com/go-chi/chi/v5"
)

type CreateRequest struct {
	ID      string    `json:"id"`
	At      time.Time `json:"at"`
	WebHook string    `json:"wh"`
}

func GetHttpServiceHandler(service *CronService) http.Handler {
	httpRoute := chi.NewRouter()

	httpRoute.Get("/", func(w http.ResponseWriter, r *http.Request) {
		output := base.ServiceRootResponse{
			ServieName: ServiceShortName,
			Version:    ServiceVersion,
			Revision:   base.GetRevision(),
		}
		body, _ := json.Marshal(output)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	httpRoute.Get("/list", func(w http.ResponseWriter, r *http.Request) {
		query := Query{}
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			base.MakeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		query.Limit = limit

		after := r.URL.Query().Get("after")
		if after != "" {
			afterTime, err := time.Parse(time.RFC3339, after)
			if err != nil {
				base.MakeErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			query.After = afterTime
		}

		list, err := service.List(query)
		base.MakeResponse(w, list, err)
	})

	httpRoute.Post("/create", func(w http.ResponseWriter, r *http.Request) {
		var req CreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			base.MakeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		err = service.Create(req.ID, req.At, req.WebHook)
		base.MakeResponse(w, err, err)
	})

	httpRoute.Get("/{recordID}/abort", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		err := service.Abort(id)
		base.MakeResponse(w, err, err)
	})

	return httpRoute
}
