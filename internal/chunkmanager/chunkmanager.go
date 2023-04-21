package chunkmanager

import (
	"errors"
	"log"
	"simple-storage/internal/utils"
	"sort"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrAlreadyExist             = errors.New("file already exist")
	ErrNotFound                 = errors.New("file not found")
	ErrNoStorageServerAvailable = errors.New("no storage server available")
)

type Chunk struct {
	ID            string
	StorageServer string
}

type storageServer struct {
	address        string
	numberOfChunks int
}

type file struct {
	chunks []Chunk
	size   int
}

type ChunkManager struct {
	log                    *log.Logger
	config                 Config
	storageServerByAddress map[string]struct{} // address
	storageServers         []storageServer
	files                  map[string]file
	sync.Mutex
}

type Config struct {
	MaxChunkSizeBytes     int
	ErasureCodingFraction int
}

func New(log *log.Logger, config Config) *ChunkManager {
	log = utils.LoggerExtendWithPrefix(log, "chunk-manager ->")

	return &ChunkManager{
		log:                    log,
		config:                 config,
		storageServerByAddress: make(map[string]struct{}),
		files:                  make(map[string]file),
	}

}

func (cm *ChunkManager) RegisterStorageServer(address string) error {
	cm.Lock()
	defer cm.Unlock()

	if _, ok := cm.storageServerByAddress[address]; !ok {
		cm.storageServerByAddress[address] = struct{}{}
		cm.storageServers = append(cm.storageServers, storageServer{
			address:        address,
			numberOfChunks: 0,
		})

		cm.log.Printf("Register new storage server %s new ss table %v",
			address, cm.storageServerByAddress)
	}

	return nil
}

func (cm *ChunkManager) SplitIntoChunks(
	filename string, filesize int,
) ([]Chunk, error) {
	cm.Lock()
	defer cm.Unlock()

	if _, ok := cm.files[filename]; ok {
		return nil, ErrAlreadyExist
	}

	if len(cm.storageServers) == 0 {
		return nil, ErrNoStorageServerAvailable
	}

	sort.Slice(cm.storageServers, func(i, j int) bool {
		return cm.storageServers[i].numberOfChunks < cm.storageServers[j].numberOfChunks
	})

	var (
		cChunk = numberOfChunks(filesize, cm.config.ErasureCodingFraction, cm.config.MaxChunkSizeBytes)
		chunks = make([]Chunk, 0, cChunk)
		j      = 0
	)

	for i := 0; i < cChunk; i++ {
		chunks = append(chunks, Chunk{
			ID:            uuid.New().String(),
			StorageServer: cm.storageServers[j].address,
		})

		cm.storageServers[j].numberOfChunks++

		j++

		if j >= len(cm.storageServers) {
			j = 0
		}
	}

	cm.files[filename] = file{chunks: chunks, size: filesize}

	cm.log.Printf("Split %s [%d] into %d chunks", filename, filesize, len(chunks))

	return chunks, nil
}

func (cm *ChunkManager) ChunksInfo(filename string) ([]Chunk, int, error) {
	cm.Lock()
	defer cm.Unlock()

	file, ok := cm.files[filename]

	if !ok {
		return nil, 0, ErrNotFound
	}

	return file.chunks, file.size, nil
}
