package base

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceRootResponse struct {
	ServieName string `json:"name"`
	Version    string `json:"ver"`
	Revision   string `json:"rev"`
}

type WebServer struct {
	server *http.Server
}

func NewWebServer(config *WebConfig, httpRoute *chi.Mux) (*WebServer, error) {
	httpServer := &http.Server{Addr: config.BindingAddress, Handler: httpRoute}
	return &WebServer{server: httpServer}, nil
}

func (w *WebServer) ListenAndServer() error {
	return w.server.ListenAndServe()
}

func (w *WebServer) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}

func MakeErrorResponse(w http.ResponseWriter, status int, errMsg string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(errMsg))
}

func MakeSuccessResponse[T any](w http.ResponseWriter, content T) {
	body, err := json.Marshal(content)
	if err != nil {
		MakeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Fail to prepare response. Error=%v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func MakeResponse[T any](w http.ResponseWriter, content T, err error) {
	if err != nil {
		MakeErrorResponse(w, http.StatusInternalServerError, err.Error())
	} else {
		MakeSuccessResponse(w, content)
	}
}
