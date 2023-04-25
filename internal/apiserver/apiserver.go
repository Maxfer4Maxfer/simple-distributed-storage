package apiserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	cm "simple-storage/internal/chunkmanager"
	"simple-storage/internal/utils"
)

var (
	ErrUploadCanceled   = errors.New("uploading has been canceled")
	ErrDownloadCanceled = errors.New("downloading has been canceled")
)

type ChunkManager interface {
	SplitIntoChunks(filename string, size int64) ([]cm.Chunk, error)
	ChunksInfo(filename string) ([]cm.Chunk, int64, error)
}

type StorageServer interface {
	UploadChunk(chunkID string, buf []byte) error
	DownloadChunk(chunkID string, buf []byte) error
}

type APIServer struct {
	log            *log.Logger
	config         Config
	cm             ChunkManager
	storageServers storageServerKeeper
}

type Config struct{}

type StorageServerClientCreatorFunc func(address string) StorageServer

func New(
	log *log.Logger, config Config,
	chunkManager ChunkManager, ssClientCreator StorageServerClientCreatorFunc,
) *APIServer {
	log = utils.LoggerExtendWithPrefix(log, "api-server ->")

	return &APIServer{
		log:    log,
		config: config,
		cm:     chunkManager,
		storageServers: storageServerKeeper{
			storageServers:                 map[string]StorageServer{},
			storageServerClientCreatorFunc: ssClientCreator,
		},
	}
}

func (s *APIServer) PutObject(
	ctx context.Context, filename string, r io.Reader, size int64,
) error {
	chunks, err := s.cm.SplitIntoChunks(filename, size)
	if err != nil {
		return fmt.Errorf("failure to split file into chunks: %w", err)
	}

	chunkSize := utils.ChunkSize(size, len(chunks))
	buf := make([]byte, chunkSize)

	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ErrUploadCanceled
		default:
		}

		ss := s.storageServers.get(chunk.StorageServer)

		n, err := r.Read(buf)
		if err != nil {
			return fmt.Errorf("failure to read filename: %s: %w ", filename, err)
		}

		err = ss.UploadChunk(chunk.ID, buf[:n])
		if err != nil {
			return fmt.Errorf("failure to upload "+
				"filename: %s chunk: %s storage-server: %s: %w ",
				filename, chunk.ID, chunk.StorageServer, err)
		}
	}

	return nil
}

func (s *APIServer) GetObject(
	ctx context.Context, filename string, w io.Writer,
) error {
	chunks, filesize, err := s.cm.ChunksInfo(filename)
	if err != nil {
		return fmt.Errorf("failure to get file's info: %w", err)
	}

	chunksize := utils.ChunkSize(filesize, len(chunks))
	buf := make([]byte, chunksize)
	restsize := int(filesize)

	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ErrDownloadCanceled
		default:
		}

		ss := s.storageServers.get(chunk.StorageServer)
		n := min(restsize, chunksize)

		err := ss.DownloadChunk(chunk.ID, buf[:n])
		if err != nil {
			return fmt.Errorf("failure to download "+
				"chunk: %s of filename: %s from storage-server: %s: %w",
				filename, chunk.ID, chunk.StorageServer, err)
		}

		_, err = io.Copy(w, bytes.NewReader(buf[:n]))
		if err != nil {
			return fmt.Errorf("failure to upload "+
				"chunk: %s of filename: %s from storage-server: %s: %w",
				filename, chunk.ID, chunk.StorageServer, err)
		}

		restsize -= n
	}

	return nil
}
