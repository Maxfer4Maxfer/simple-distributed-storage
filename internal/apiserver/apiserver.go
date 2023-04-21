package apiserver

import (
	"fmt"
	"log"
	cm "simple-storage/internal/chunkmanager"
	"simple-storage/internal/utils"
)

type ChunkManager interface {
	SplitIntoChunks(filename string, size int) ([]cm.Chunk, error)
	ChunksInfo(filename string) ([]cm.Chunk, int, error)
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

func (s *APIServer) PutObject(filename string, buf []byte) error {
	chunks, err := s.cm.SplitIntoChunks(filename, len(buf))
	if err != nil {
		return fmt.Errorf("failure to split file into chunks: %w", err)
	}

	chunkSize := utils.ChunkSize(len(buf), len(chunks))
	offset := 0

	for i, chunk := range chunks {
		ss := s.storageServers.get(chunk.StorageServer)

		offset = i * chunkSize

		err := ss.UploadChunk(chunk.ID, buf[offset:min(offset+chunkSize, len(buf))])
		if err != nil {
			return fmt.Errorf("failure to upload "+
				"filename: %s chunk: %s storage-server: %s: %w ",
				filename, chunk.ID, chunk.StorageServer, err)
		}
	}

	return nil
}

func (s *APIServer) GetObject(filename string) ([]byte, error) {
	chunks, filesize, err := s.cm.ChunksInfo(filename)
	if err != nil {
		return nil, fmt.Errorf("failure to get file's info: %w", err)
	}

	buf := make([]byte, filesize)
	chunkSize := utils.ChunkSize(filesize, len(chunks))
	offset := 0

	for i, chunk := range chunks {
		ss := s.storageServers.get(chunk.StorageServer)

		offset = i * chunkSize

		err := ss.DownloadChunk(chunk.ID, buf[offset:min(offset+chunkSize, len(buf))])
		if err != nil {
			return nil, fmt.Errorf("failure to download "+
				"filename: %s chunk: %s storage-server: %s: %w ",
				filename, chunk.ID, chunk.StorageServer, err)
		}
	}

	return buf, nil
}
