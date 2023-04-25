package apiserver

import (
	"bytes"
	"log"
	"simple-storage/internal/chunkmanager"
	"simple-storage/tests/mock"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAPIServer_PutObject(t *testing.T) {
	tt := []struct {
		filename string
		chunks   []chunkmanager.Chunk
		buf      string
	}{
		{
			filename: "file1",
			chunks: []chunkmanager.Chunk{
				{ID: "id1", StorageServer: "0.0.0.0:9001"},
				{ID: "id2", StorageServer: "0.0.0.0:9002"},
			},
			buf: "Hello World!",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range tt {
		cm := mock.NewMockChunkManager(ctrl)
		cm.EXPECT().SplitIntoChunks(tc.filename, int64(len(tc.buf))).
			Return(tc.chunks, nil).Times(1)

		ssClientCreator := func(_ string) StorageServer {
			ss := mock.NewMockStorageServer(ctrl)
			ss.EXPECT().UploadChunk(gomock.Any(), gomock.Any()).Return(nil).Times(1)

			return ss
		}

		apiserver := New(log.Default(), Config{}, cm, ssClientCreator)

		r := strings.NewReader(tc.buf)
		err := apiserver.PutObject(tc.filename, r, int64(len(tc.buf)))
		require.NoError(t, err)
	}
}

func TestAPIServer_GetObject(t *testing.T) {
	tt := []struct {
		filename   string
		filesize   int64
		chunks     []chunkmanager.Chunk
		ssResponce map[string][]byte
		result     string
	}{
		{
			filename: "file1",
			filesize: 12,
			chunks: []chunkmanager.Chunk{
				{ID: "chunkID1", StorageServer: "0.0.0.0:9001"},
				{ID: "chunkID2", StorageServer: "0.0.0.0:9002"},
			},
			ssResponce: map[string][]byte{
				"chunkID1": []byte("Hello "),
				"chunkID2": []byte("World!"),
			},
			result: "Hello World!",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range tt {
		cm := mock.NewMockChunkManager(ctrl)
		cm.EXPECT().ChunksInfo(tc.filename).
			Return(tc.chunks, tc.filesize, nil).Times(1)

		ssClientCreator := func(_ string) StorageServer {
			ss := mock.NewMockStorageServer(ctrl)
			ss.EXPECT().DownloadChunk(gomock.Any(), gomock.Any()).
				DoAndReturn(
					func(id string, buf []byte) error {
						res := tc.ssResponce[id]
						copy(buf, res)
						return nil
					},
				).Times(1)

			return ss
		}

		apiserver := New(log.Default(), Config{}, cm, ssClientCreator)

		buf := new(bytes.Buffer)

		err := apiserver.GetObject(tc.filename, buf)
		require.NoError(t, err)
		require.Equal(t, buf.String(), tc.result)
	}
}
