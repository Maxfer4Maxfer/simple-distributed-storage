package chunkmanager

import (
	"math"
)

func numberOfChunks(filesize, erasureCodingFraction, maxChunkSize int) int {
	chunkSize := int(math.Ceil(float64(filesize) / float64(erasureCodingFraction)))

	if chunkSize < maxChunkSize {
		return int(math.Ceil(float64(filesize) / float64(chunkSize)))
	} else {
		return int(math.Ceil(float64(filesize) / float64(maxChunkSize)))
	}
}
