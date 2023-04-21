package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	lhttp "simple-storage/internal/entrypoint/http"
	"simple-storage/internal/utils"
)

type StorageServer interface {
	UploadChunk(chunkID string, file io.Reader) error
	DownloadChunk(chunkID string) ([]byte, error)
}

// Handler is a wraper on http.Server.
type Handler struct {
	log           *log.Logger
	storageServer StorageServer
	*lhttp.Handler
}

// New returns a HTTP server.
func New(
	log *log.Logger,
	storageServer StorageServer,
) *Handler {
	log = utils.LoggerExtendWithPrefix(log, "http-handler ->")

	return &Handler{
		log,
		storageServer,
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
		case r.URL.Path == "/" && r.Method == http.MethodPut:
			han.handleUpload().ServeHTTP(w, r)
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

		buf, err := han.storageServer.DownloadChunk(chunkID[0])
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

		file, header, err := r.FormFile("chunk")
		if err != nil {
			han.ResponseWithError(w, r, err)
			return
		}

		defer file.Close()

		err = han.storageServer.UploadChunk(header.Filename, file)
		if err != nil {
			han.ResponseWithError(w, r, err)

			return
		}

		han.HandleOK().ServeHTTP(w, r)
	})
}
