package utils

import "math"

func ChunkSize(fileSize int, cChunk int) int {
	return int(math.Ceil(float64(fileSize) / float64(cChunk)))
}

