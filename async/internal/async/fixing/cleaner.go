package fixing

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/async/internal/async/config"
	"github.com/UnicomAI/wanwu/async/internal/db/client"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
)

// IClean goroutine safe
type IClean interface {
	Run(ctx context.Context) error
	Stop()
}

type cleaner struct {
	log    async_config.Logger
	client client.ITaskClient

	cleanTimeout int // second

	cleanInterval int // second

	mutex   sync.Mutex
	stopped bool
	stop    chan struct{}

	wg sync.WaitGroup
}

func NewClean(c client.ITaskClient, cfg config.Config) IClean {
	return &cleaner{
		log:           cfg.Log,
		client:        c,
		cleanTimeout:  config.CleanTimeout,
		cleanInterval: config.CleanInterval,
		stop:          make(chan struct{}, 1),
	}
}

func (c *cleaner) Run(ctx context.Context) error {
	c.mutex.Lock()
	if c.stopped {
		defer c.mutex.Unlock()
		return errors.New("async cleaner already stop")
	}
	c.wg.Add(1)
	c.mutex.Unlock()

	go func() {
		defer tools.PrintPanicStack()
		defer c.wg.Done()

		c.log.Infof("async cleaner run")
		// clean once when start
		if err := c.client.Clean(ctx, c.cleanTimeout); err != nil {
			c.log.Errorf("async cleaner err: %v", err)
		}
		ticker := time.NewTicker(time.Duration(c.cleanInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			case <-ticker.C:
				if err := c.client.Clean(ctx, c.cleanTimeout); err != nil {
					c.log.Errorf("async cleaner err: %v", err)
				}
			}
		}
	}()
	return nil
}

func (c *cleaner) Stop() {
	c.mutex.Lock()
	// check stop
	if c.stopped {
		defer c.mutex.Unlock()
		c.log.Errorf("async cleaner already stop")
		return
	}
	// stop
	c.stopped = true
	c.stop <- struct{}{}
	// wait
	c.wg.Wait()
	c.log.Infof("async cleaner stop")
}
