package pool

import (
	"github.com/panjf2000/ants"
	"log"
)

func NewExecutorPool(size int) *ants.Pool {
	pool, err := ants.NewPool(size)
	if err != nil {
		log.Fatalf("🔔🔔🔔 fatal error NewPool: %v 🔔🔔🔔", err)
	}
	return pool
}
