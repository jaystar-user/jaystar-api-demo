package utils

import (
	"errors"
	"math"
)

func RunInBatch(total int, batchSize int, do func(batchIndex int, start int, end int) error) (err error) {
	if batchSize == 0 {
		return errors.New("invalid batchSize")
	}
	batchCount := getBatchCount(total, batchSize)
	for i := 0; i < batchCount; i++ {
		batchIndex := i * batchSize
		var end int
		if i == batchCount-1 {
			end = total
		} else {
			end = batchIndex + batchSize
		}

		err = do(i+1, batchIndex, end)
		if err != nil {
			return
		}
	}

	return
}

func getBatchCount(total int, batchSize int) int {
	return int(math.Ceil(float64(total) / float64(batchSize)))
}
