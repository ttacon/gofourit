package gofourit

import (
	"github.com/robfig/cron/v3"
)

// Cron is the interface for our distributed cron scheduler
type Cron interface {
	AddFunc(cronString, key string, fn func()) Cron
	Entries() []cron.Entry

	Start()
	Stop()
}

// RemoteSource is the source of how we acquire locks to run jobs on a single node.
type RemoteSource interface {
	GrabLock(key string) Lock
}

// Lock is the interface for Locks generated by our RemoteSource.
type Lock interface {
	Release()
}

// cronImpl is the basic implementation which mirrors the basic API from
// robfig/cron and wraps functions in with a bit of boilerplate that ensures
// only a single node with this registered function executes.
type cronImpl struct {
	cron *cron.Cron
	rSrc RemoteSource
}

// New returns a new Cron that will use the given RemoteSource to generate
// generate locks.
func New(rSrc RemoteSource) Cron {
	return &cronImpl{
		cron: cron.New(),
		rSrc: rSrc,
	}
}

// AddFunc adds the given func to be run on the given schedule, using the given
// key to ensure only one copy is ever run at a time.
func (c *cronImpl) AddFunc(cronString, key string, fn func()) Cron {
	c.cron.AddFunc(cronString, func() {
		lock := c.rSrc.GrabLock(key)
		if lock == nil {
			// We didn't get the lock, exit.
			return
		}
		// Use defer here instead of calling below incase `fn` panics.
		defer lock.Release()

		fn()
	})
	return c
}

// Entries return our registered cron entries.
func (c *cronImpl) Entries() []cron.Entry {
	return c.cron.Entries()
}

// Start starts the cron scheduler.
func (c *cronImpl) Start() {
	c.cron.Start()
}

// Stop stops the cron scheduler.
func (c *cronImpl) Stop() {
	c.cron.Stop()
}
