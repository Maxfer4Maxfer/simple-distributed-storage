package storageserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"simple-storage/internal/utils"
	"sync"
	"time"
)

type ChunkManager interface {
	RegisterStorageServer(address string) error
}

type StorageServer struct {
	log    *log.Logger
	config Config
	sync.Mutex
}

type Config struct {
	Address                            string
	DataDirectory                      string
	TimeBetweetRegistrationRetrySecond int
}

func New(
	log *log.Logger, config Config, cm ChunkManager,
) *StorageServer {
	log = utils.LoggerExtendWithPrefix(log, "storage-server ->")

	ss := &StorageServer{
		log:    log,
		config: config,
	}

	go ss.Register(cm)

	return ss
}

func (ss *StorageServer) Register(cm ChunkManager) {
	for {
		if err := cm.RegisterStorageServer(ss.config.Address); err != nil {
			ss.log.Printf("ERROR: failure to register itself: %s", err)

			time.Sleep(time.Duration(
				ss.config.TimeBetweetRegistrationRetrySecond) * time.Second)

			continue
		}

		return
	}
}

func (ss *StorageServer) UploadChunk(chunkID string, in io.Reader) error {
	file, err := os.Create(filepath.Join(ss.config.DataDirectory, chunkID))
	if err != nil {
		return fmt.Errorf("failure to save chunk: %w", err)
	}

	_, err = io.Copy(file, in)
	if err != nil {
		return fmt.Errorf("failure to save chunk: %w", err)
	}

	return nil
}

func (ss *StorageServer) DownloadChunk(chunkID string) ([]byte, error) {
	return  ioutil.ReadFile(filepath.Join(ss.config.DataDirectory, chunkID))
}
