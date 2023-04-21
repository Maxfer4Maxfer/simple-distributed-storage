package chunkmanager

import (
	"log"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkManager_RegisterStorageServer(t *testing.T) {
	tt := []struct {
		addresses []string
		result    map[string]struct{}
	}{
		{
			addresses: []string{"0.0.0.0:9091"},
			result:    map[string]struct{}{"0.0.0.0:9091": struct{}{}},
		},
		{
			addresses: []string{
				"0.0.0.0:9091",
				"0.0.0.0:9092",
				"0.0.0.0:9091",
			},
			result: map[string]struct{}{
				"0.0.0.0:9091": struct{}{},
				"0.0.0.0:9092": struct{}{},
			},
		},
	}

	cm := New(log.Default(), Config{})

	for _, tc := range tt {
		for _, address := range tc.addresses {
			err := cm.RegisterStorageServer(address)
			require.NoError(t, err)
		}
		require.Equal(t, tc.result, cm.storageServerByAddress)
	}
}

func TestChunkManager_SplitIntoChunks_SimpleDistribution(t *testing.T) {
	tt := []struct {
		storageServers        []string
		filename              string
		filesize              int
		maxChunkSizeBytes     int
		erasureCodingFraction int
		cChunk                int
		distributionChunk     map[string]int
	}{
		{
			storageServers: []string{
				"0.0.0.0:9091",
				"0.0.0.0:9092",
			},
			maxChunkSizeBytes:     int(math.MaxInt64),
			erasureCodingFraction: 2,
			filename:              "file1",
			filesize:              100,
			cChunk:                2,
			distributionChunk: map[string]int{
				"0.0.0.0:9091": 1,
				"0.0.0.0:9092": 1,
			},
		},
		{
			storageServers: []string{
				"0.0.0.0:9091",
				"0.0.0.0:9092",
			},
			maxChunkSizeBytes:     int(math.MaxInt64),
			erasureCodingFraction: 2,
			filename:              "file1",
			filesize:              101,
			cChunk:                2,
			distributionChunk: map[string]int{
				"0.0.0.0:9091": 1,
				"0.0.0.0:9092": 1,
			},
		},
		{
			storageServers: []string{
				"0.0.0.0:9091",
				"0.0.0.0:9092",
			},
			maxChunkSizeBytes:     50,
			erasureCodingFraction: 2,
			filename:              "file1",
			filesize:              101,
			cChunk:                3,
			distributionChunk: map[string]int{
				"0.0.0.0:9091": 2,
				"0.0.0.0:9092": 1,
			},
		},
		{
			storageServers: []string{
				"0.0.0.0:9091",
				"0.0.0.0:9092",
			},
			maxChunkSizeBytes:     25,
			erasureCodingFraction: 2,
			filename:              "file1",
			filesize:              100,
			cChunk:                4,
			distributionChunk: map[string]int{
				"0.0.0.0:9091": 2,
				"0.0.0.0:9092": 2,
			},
		},
	}

	for _, tc := range tt {
		cm := New(log.Default(), Config{
			MaxChunkSizeBytes:     tc.maxChunkSizeBytes,
			ErasureCodingFraction: tc.erasureCodingFraction,
		})

		for _, ss := range tc.storageServers {
			err := cm.RegisterStorageServer(ss)
			require.NoError(t, err)
		}

		chunks, err := cm.SplitIntoChunks(tc.filename, tc.filesize)
		require.NoError(t, err)
		require.Equal(t, tc.cChunk, len(chunks))

		distribution := make(map[string]int, len(chunks))
		for _, chunk := range chunks {
			distribution[chunk.StorageServer]++
		}
		require.Equal(t, tc.distributionChunk, distribution)
	}
}

func TestChunkManager_SplitIntoChunks_ConsistentDistribution(t *testing.T) {
	tt := []struct {
		storageServers        []string
		firstFilename         string
		firstFilesize         int
		secondFilename        string
		secondFilesize        int
		maxChunkSizeBytes     int
		erasureCodingFraction int
		cChunk                int
		firstDistributionChunk     map[string]int
		secondDistributionChunk     map[string]int
	}{
		{
			storageServers: []string{
				"0.0.0.0:9091",
				"0.0.0.0:9092",
				"0.0.0.0:9093",
			},
			maxChunkSizeBytes:     int(math.MaxInt64),
			erasureCodingFraction: 2,
			firstFilename:         "file1",
			firstFilesize:         100,
			firstDistributionChunk: map[string]int{
				"0.0.0.0:9091": 1,
				"0.0.0.0:9092": 1,
			},
			secondFilename: "file2",
			secondFilesize: 100,
			secondDistributionChunk: map[string]int{
				"0.0.0.0:9091": 1,
				"0.0.0.0:9093": 1,
			},
		},
	}

	for _, tc := range tt {
		cm := New(log.Default(), Config{
			MaxChunkSizeBytes:     tc.maxChunkSizeBytes,
			ErasureCodingFraction: tc.erasureCodingFraction,
		})

		for _, ss := range tc.storageServers {
			err := cm.RegisterStorageServer(ss)
			require.NoError(t, err)
		}

		chunks, err := cm.SplitIntoChunks(tc.firstFilename, tc.firstFilesize)
		require.NoError(t, err)

		distribution := make(map[string]int, len(chunks))
		for _, chunk := range chunks {
			distribution[chunk.StorageServer]++
		}
		require.Equal(t, tc.firstDistributionChunk, distribution)

		chunks, err = cm.SplitIntoChunks(tc.secondFilename, tc.secondFilesize)
		require.NoError(t, err)

		distribution = make(map[string]int, len(chunks))
		for _, chunk := range chunks {
			distribution[chunk.StorageServer]++
		}
		require.Equal(t, tc.secondDistributionChunk, distribution)
	}
}
