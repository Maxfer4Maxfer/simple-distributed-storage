package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	lhttp "simple-storage/internal/entrypoint/http"
	"simple-storage/internal/utils"
)

type APIServer interface {
	PutObject(filename string, buf []byte) error
	GetObject(filename string) ([]byte, error)
}

type ChunkManager interface {
	RegisterStorageServer(address string) error
}

// Handler is a wraper on http.Server.
type Handler struct {
	log          *log.Logger
	apiServer    APIServer
	chunkManager ChunkManager
	*lhttp.Handler
}

// New returns a HTTP server.
func New(
	log *log.Logger,
	apiServer APIServer,
	chunkManager ChunkManager,
) *Handler {
	log = utils.LoggerExtendWithPrefix(log, "http-handler ->")

	return &Handler{
		log,
		apiServer,
		chunkManager,
		lhttp.NewHandler(log),
	}
}

// ServeHTTP configures and returns a new router.
func (han *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodOptions:
			han.HandleOK().ServeHTTP(w, r)
		case r.URL.Path == "/" && r.Method == http.MethodGet:
			han.handleDownload().ServeHTTP(w, r)
		case filepath.Dir(r.URL.Path) == "/" && r.Method == http.MethodPut:
			han.handleUpload().ServeHTTP(w, r)
		case r.URL.Path == "/register" && r.Method == http.MethodPost:
			han.handleRegister().ServeHTTP(w, r)
		default:
			han.HandleEmpty().ServeHTTP(w, r)
		}
	})

	han.HandleCORS(router).ServeHTTP(w, r)
}

func (han *Handler) handleDownload() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chunkID, ok := r.URL.Query()["id"]
		if !ok {
			han.ResponseWithError(w, r, errors.New("id should be set"))
			return
		}

		buf, err := han.apiServer.GetObject(chunkID[0])
		if err != nil {
			han.ResponseWithError(w, r, err)

			return
		}

		han.ResponseWithData(w, r, buf)
	})
}

func (han *Handler) handleUpload() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)

		file, header, err := r.FormFile("file")
		if err != nil {
			han.ResponseWithError(w, r, err)
			return
		}
		defer file.Close()

		buf := make([]byte, header.Size)

		_, err = file.Read(buf)
		if err != nil {
			han.ResponseWithError(w, r,
				fmt.Errorf("cannot read incomming file: %w", err))
			return
		}

		err = han.apiServer.PutObject(header.Filename, buf)
		if err != nil {
			han.ResponseWithError(w, r, err)

			return
		}

		han.HandleOK().ServeHTTP(w, r)
	})
}

func (han *Handler) handleRegister() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			han.ResponseWithError(w, r, err)

			return
		}

		address := string(body)

		if len(address) == 0 {
			han.ResponseWithError(w, r, errors.New("address should be set"))
			return
		}

		err = han.chunkManager.RegisterStorageServer(address)
		if err != nil {
			han.ResponseWithError(w, r, err)

			return
		}

		han.HandleOK().ServeHTTP(w, r)
	})
}
