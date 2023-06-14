package bgjobs

import (
	"context"
	"sync"

	"github.com/petuhovskiy/neon-lights/internal/log"
)

// Register is a registry of all background jobs.
// Has a wait group to wait for all jobs to finish.
type Register struct {
	all sync.WaitGroup
}

func NewRegister() *Register {
	return &Register{}
}

// Go a new background task.
func (r *Register) Go(f func()) {
	r.all.Add(1)

	go func() {
		defer r.all.Done()
		f()
	}()
}

func (r *Register) WaitAll(ctx context.Context) {
	log.Info(ctx, "waiting for all background jobs to finish")
	r.all.Wait()
}
