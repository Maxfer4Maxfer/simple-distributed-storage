package handler

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"simple-storage/internal/apiserver"
	lhttp "simple-storage/internal/entrypoint/http"
	"simple-storage/internal/utils"
)

type APIServer interface {
	PutObject(ctx context.Context, filename string, r io.Reader, size int64) error
	GetObject(ctx context.Context, filename string, w io.Writer) error
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

const (
	StatusClientClosedRequest = 499
)

func (han *Handler) handleDownload() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chunkID, ok := r.URL.Query()["id"]
		if !ok {
			han.ResponseWithError(
				w, r, errors.New("id should be set"), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		err := han.apiServer.GetObject(ctx, chunkID[0], w)
		if err != nil {
			if errors.Is(err, apiserver.ErrDownloadCanceled) {
				han.ResponseWithError(w, r, err, StatusClientClosedRequest)
			} else {
				han.ResponseWithError(w, r, err, http.StatusInternalServerError)
			}

			return
		}
	})
}

func (han *Handler) handleUpload() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)

		file, header, err := r.FormFile("file")
		if err != nil {
			han.ResponseWithError(w, r, err, http.StatusBadRequest)
			return
		}
		defer file.Close()

		ctx := r.Context()

		err = han.apiServer.PutObject(ctx, header.Filename, file, header.Size)
		if err != nil {
			if errors.Is(err, apiserver.ErrUploadCanceled) {
				han.ResponseWithError(w, r, err, StatusClientClosedRequest)
			} else {
				han.ResponseWithError(w, r, err, http.StatusInternalServerError)
			}

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
			han.ResponseWithError(w, r, err, http.StatusBadRequest)

			return
		}

		address := string(body)

		if len(address) == 0 {
			han.ResponseWithError(
				w, r, errors.New("address should be set"), http.StatusBadRequest)
			return
		}

		err = han.chunkManager.RegisterStorageServer(address)
		if err != nil {
			han.ResponseWithError(w, r, err, http.StatusInternalServerError)

			return
		}

		han.HandleOK().ServeHTTP(w, r)
	})
}
