package utils

import "math"

func ChunkSize(filesize int64, cChunk int) int {
	return int(math.Ceil(float64(filesize) / float64(cChunk)))
}

