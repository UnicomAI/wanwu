package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/redis/go-redis/v9"
)

var redisClient = RedisClient{}

type RedisClient struct {
	Client       redis.UniversalClient
	trimTicker   *time.Ticker
	trimStopChan chan bool
	trimRunning  bool
}

func init() {
	pkg.AddContainer(redisClient)
}

func (c RedisClient) LoadType() string {
	return "redis"
}

func (c RedisClient) Load() error {
	if !config.GetConfig().Redis.Enabled {
		log.Infof("Redis is not enabled, skip init")
		return nil
	}

	log.Infof("Redis is enabled, start init")
	client, err := initRedisClient()
	if err != nil {
		return err
	}
	redisClient.Client = client

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Errorf("Redis连接测试失败: %v", err)
		return err
	}

	log.Infof("Redis客户端初始化成功")
	startTrimTask()

	return nil
}

func (c RedisClient) Stop() error {
	if !config.GetConfig().Redis.Enabled {
		return nil
	}

	stopTrimTask()

	if redisClient.Client != nil {
		return redisClient.Client.Close()
	}
	return nil
}

func (c RedisClient) StopPriority() int {
	return pkg.DefaultPriority
}

func initRedisClient() (redis.UniversalClient, error) {
	cfg := config.GetConfig().Redis
	log.Infof("开始初始化Redis客户端，模式: %s", cfg.Mode)

	var client redis.UniversalClient
	var err error

	switch cfg.Mode {
	case "standalone":
		client, err = initStandaloneClient(cfg)
	case "sentinel":
		client, err = initSentinelClient(cfg)
	case "cluster":
		client, err = initClusterClient(cfg)
	default:
		return nil, fmt.Errorf("不支持的Redis模式: %s，支持的模式: standalone, sentinel, cluster", cfg.Mode)
	}

	if err != nil {
		log.Errorf("Redis客户端初始化失败: %v", err)
		return nil, err
	}

	log.Infof("Redis客户端创建成功，模式: %s", cfg.Mode)
	return client, nil
}

func initStandaloneClient(cfg *config.RedisConfig) (*redis.Client, error) {
	if len(cfg.Addr) == 0 {
		return nil, errors.New("standalone模式需要至少一个地址")
	}

	// 取第一个地址作为standalone地址
	addr := cfg.Addr[0]

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		ConnMaxIdleTime: time.Duration(cfg.IdleTimeout) * time.Second,
	})

	return client, nil
}

func initSentinelClient(cfg *config.RedisConfig) (*redis.Client, error) {
	if len(cfg.Addr) == 0 {
		return nil, errors.New("sentinel模式需要至少一个哨兵地址")
	}
	if cfg.MasterName == "" {
		return nil, errors.New("sentinel模式需要指定master-name")
	}

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      cfg.MasterName,
		SentinelAddrs:   cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		ConnMaxIdleTime: time.Duration(cfg.IdleTimeout) * time.Second,
	})

	return client, nil
}

func initClusterClient(cfg *config.RedisConfig) (*redis.ClusterClient, error) {
	if len(cfg.Addr) == 0 {
		return nil, errors.New("cluster模式需要至少一个节点地址")
	}

	// 注意：cluster模式下DB参数无效，Redis集群只支持DB 0
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           cfg.Addr,
		Password:        cfg.Password,
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		ConnMaxIdleTime: time.Duration(cfg.IdleTimeout) * time.Second,
	})

	return client, nil
}

// sendMessageToRedis 发送消息到Redis Stream
func sendMessageToRedis(msg interface{}, streamKey string) error {
	if msg == nil {
		return errors.New("message is nil")
	}

	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// 使用XAdd命令将消息添加到Stream
	// 格式: XADD streamKey * field value
	// * 表示让Redis自动生成消息ID
	result, err := redisClient.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data":      message,
			"timestamp": time.Now().Unix(),
		},
	}).Result()

	if err != nil {
		log.Errorf("Redis Stream发送消息失败, stream: %s, error: %v", streamKey, err)
		return err
	}

	log.Infof("Redis Stream发送成功, stream: %s, messageId: %s, data: %s",
		streamKey, result, message)
	return nil
}

// startTrimTask 启动定时修剪任务
func startTrimTask() {
	cfg := config.GetConfig().Redis

	// 检查是否需要启用定时修剪
	if cfg.StreamMaxLen <= 0 {
		log.Infof("StreamMaxLen配置为0或负数，不启用定时修剪任务")
		return
	}

	// 初始化通道和ticker
	redisClient.trimStopChan = make(chan bool)
	redisClient.trimRunning = true

	// 启动定时任务（每分钟执行一次）
	redisClient.trimTicker = time.NewTicker(1 * time.Minute)

	go func() {
		log.Infof("Redis Stream定时修剪任务已启动")

		for {
			select {
			case <-redisClient.trimTicker.C:
				// 执行修剪任务
				performTrimTask()
			case <-redisClient.trimStopChan:
				log.Infof("Redis Stream定时修剪任务已停止")
				return
			}
		}
	}()
}

// stopTrimTask 停止定时修剪任务
func stopTrimTask() {
	if redisClient.trimRunning {
		redisClient.trimRunning = false

		// 停止ticker
		if redisClient.trimTicker != nil {
			redisClient.trimTicker.Stop()
		}

		// 发送停止信号
		if redisClient.trimStopChan != nil {
			redisClient.trimStopChan <- true
			close(redisClient.trimStopChan)
		}
	}
}

// performTrimTask 执行修剪任务
func performTrimTask() {
	cfg := config.GetConfig().Redis

	// 获取配置中的Stream最大长度
	maxLen := cfg.StreamMaxLen
	if maxLen <= 0 {
		return
	}

	ctx := context.Background()

	topics := []string{
		config.GetConfig().Topic.UrlAnalysisTopic,
		config.GetConfig().Topic.UrlImportTopic,
		config.GetConfig().Topic.Topic,
		config.GetConfig().Topic.KnowledgeGraphTopic,
	}
	// 根据配置决定是否使用近似修剪
	var err error
	if cfg.StreamApproxMaxLen {
		// 使用MAXLEN近似修剪，性能更好
		for _, topic := range topics {
			_, err = redisClient.Client.XTrimMaxLenApprox(ctx, topic, maxLen, 0).Result()
			if err != nil {
				log.Warnf("Failed to trim Redis stream: %v", err)
			}
		}
	} else {
		// 使用精确修剪
		for _, topic := range topics {
			_, err = redisClient.Client.XTrimMaxLen(ctx, topic, maxLen).Result()
			if err != nil {
				log.Warnf("Failed to trim Redis stream: %v", err)
			}
		}
	}

}
