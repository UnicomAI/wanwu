package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host       string `json:"host" mapstructure:"host"`
	Port       string `json:"port" mapstructure:"port"`
	Username   string `json:"username" mapstructure:"username"`
	Password   string `json:"password" mapstructure:"password"`
	Channel    string `json:"channel" mapstructure:"channel"`
	Standalone bool   `json:"standalone" mapstructure:"standalone"`
	MasterName string `json:"master_name" mapstructure:"master_name"`
}

type SubscribeHandle func(ctx context.Context, msg *redis.Message) error

type client struct {
	ctx context.Context
	cli *redis.Client

	mutex      sync.Mutex
	subscribes map[string]struct{} // channel -> handle

	wg      sync.WaitGroup
	stopped bool
	stop    chan struct{}
}

func newClient(ctx context.Context, c Config, db int) (*client, error) {
	redisAddr := c.Host + ":" + c.Port
	var r *redis.Client
	if c.Standalone {
		r = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Username: c.Username,
			Password: c.Password,
			DB:       db,
		})
	} else {
		r = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       c.MasterName,
			SentinelAddrs:    []string{redisAddr}, // 哨兵节点地址
			DB:               db,                  // 使用的Redis数据库编号
			SentinelPassword: c.Password,
			Password:         c.Password,
		})
	}
	if _, err := r.Ping(ctx).Result(); err != nil {
		return nil, err
	}
	return &client{
		ctx:        ctx,
		cli:        r,
		subscribes: make(map[string]struct{}),
		stop:       make(chan struct{}, 1),
	}, nil
}

func (c *client) Stop() {
	c.mutex.Lock()
	if c.stopped {
		log.Errorf("redis db(%v) client already stop", c.cli.Options().DB)
		c.mutex.Unlock()
		return
	}
	c.stopped = true
	close(c.stop)
	c.mutex.Unlock()
	c.wg.Wait()
	log.Infof("redis db(%v) client stop", c.cli.Options().DB)
}

func (c *client) Cli() *redis.Client {
	return c.cli
}

// --- Generic ---

func (c *client) SetEx(ctx context.Context, key string, value interface{}, expire time.Duration) (string, error) {
	return c.cli.SetEx(ctx, key, value, expire).Result()
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	return c.cli.Get(ctx, key).Result()
}

func (c *client) Del(ctx context.Context, key string) error {
	return c.cli.Del(ctx, key).Err()
}

func (c *client) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return c.cli.Eval(ctx, script, keys, args...).Result()
}

func (c *client) Expire(ctx context.Context, key string, expire time.Duration) error {
	return c.cli.Expire(ctx, key, expire).Err()
}

// IncrWithExpire 原子自增并设置过期时间
func (c *client) IncrWithExpire(ctx context.Context, key string, expire time.Duration) (int64, error) {
	luaScript := `
		local count = redis.call("INCR", KEYS[1])
		redis.call("EXPIRE", KEYS[1], ARGV[1])
		return count
	`
	return c.cli.Eval(ctx, luaScript, []string{key}, int(expire.Seconds())).Int64()
}

// RedisTTL 获取key的剩余过期时间
func (c *client) RedisTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.cli.TTL(ctx, key).Result()
}

// --- Hash ---

type HashItem struct {
	K string
	V string
}

func (c *client) HSet(ctx context.Context, key string, items []HashItem) error {
	if len(items) == 0 {
		return nil
	}
	values := make([]string, 0, len(items)*2)
	for _, item := range items {
		values = append(values, item.K, item.V)
	}
	return c.cli.HSet(ctx, key, values).Err()
}

func (c *client) HGet(ctx context.Context, key, field string) (*HashItem, error) {
	value, err := c.cli.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return &HashItem{K: field, V: value}, nil
}

func (c *client) HGetAll(ctx context.Context, key string) ([]HashItem, error) {
	kvs, err := c.cli.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, nil
	}
	ret := make([]HashItem, 0, len(kvs))
	for k, v := range kvs {
		ret = append(ret, HashItem{K: k, V: v})
	}
	return ret, nil
}

// --- Pub/Sub ---

func (c *client) Publish(ctx context.Context, channel string, msg string) error {
	if err := c.cli.Publish(ctx, channel, msg).Err(); err != nil {
		return fmt.Errorf("redis channel %v publish msg %+v err: %v", channel, msg, err)
	}
	log.Infof("redis channel %v publish msg %v", channel, msg)
	return nil
}

func (c *client) RegisterSubscribe(channel string, handle SubscribeHandle) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.stopped {
		return fmt.Errorf("redis client already stop")
	}

	if _, ok := c.subscribes[channel]; ok {
		return fmt.Errorf("redis channel %v already subscribed", channel)
	}
	sub := c.cli.Subscribe(c.ctx, channel)
	if _, err := sub.Receive(c.ctx); err != nil {
		if closeErr := sub.Close(); closeErr != nil {
			log.Errorf("redis channel %v subscribe err: %v, close err: %v", channel, err, closeErr)
		}
		return fmt.Errorf("redis channel %v subscribe err: %v", channel, err)
	}
	c.subscribes[channel] = struct{}{}

	c.wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer c.wg.Done()
		log.Infof("redis channel %v start run", channel)
		ch := sub.Channel()
		for {
			select {
			case <-c.stop:
				c.mutex.Lock()
				delete(c.subscribes, channel)
				if err := sub.Close(); err != nil {
					log.Errorf("redis channel %v close err: %v", channel, err)
				}
				c.mutex.Unlock()
				log.Infof("redis channel %v stop", channel)
				return
			case msg, ok := <-ch:
				if !ok {
					c.mutex.Lock()
					delete(c.subscribes, channel)
					log.Errorf("redis channel %v close by remote", channel)
					c.mutex.Unlock()
					break
				}
				log.Infof("redis channel %v recv msg: %v", channel, msg)
				if msg.Channel == channel {
					if err := wrapperPanic(c.ctx, msg, handle); err != nil {
						log.Errorf("redis channel %v recv msg: %v handle err: %v", channel, msg, err)
					}
				}
			}
		}
	}()
	return nil
}

// wrapperPanic panic -> error
func wrapperPanic(ctx context.Context, msg *redis.Message, f SubscribeHandle) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	if f != nil {
		return f(ctx, msg)
	}
	return nil
}
