package mq

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"

	"github.com/IBM/sarama"
)

var kafka = Kafka{}

type Kafka struct {
	KafkaProducer sarama.SyncProducer
}

func init() {
	pkg.AddContainer(kafka)
}

func (c Kafka) LoadType() string {
	return "kafka"
}

func (c Kafka) Load() error {
	if !config.GetConfig().Kafka.Enabled {
		log.Infof("Kafka is not enabled, skip init")
		return nil
	}

	log.Infof("Kafka is enabled, start init")
	admin, err := initKafkaAdmin()
	if err != nil {
		return err
	}
	producer, err := initKafka(admin)
	if err != nil {
		return err
	}
	kafka.KafkaProducer = producer
	return nil
}

func (c Kafka) Stop() error {
	return nil
}

func (c Kafka) StopPriority() int {
	return pkg.DefaultPriority
}

func initKafkaAdmin() (sarama.ClusterAdmin, error) {
	log.Infof("开始创建新的Kafka Admin客户端")
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.MaxVersion
	kafkaConfig.Net.SASL.Enable = true                                // 启用SASL认证
	kafkaConfig.Net.SASL.User = config.GetConfig().Kafka.User         // Kafka认证用户名
	kafkaConfig.Net.SASL.Password = config.GetConfig().Kafka.Password // Kafka认证密码
	kafkaConfig.Net.SASL.Handshake = true                             // 启用SASL握手
	admin, err := sarama.NewClusterAdmin([]string{config.GetConfig().Kafka.Addr}, kafkaConfig)
	if err != nil {
		log.Errorf("创建Kafka Admin客户端失败: %v", err)
		return nil, err
	}
	log.Infof("Kafka Admin客户端创建成功")
	return admin, nil
}

func initKafka(kafkaAdmin sarama.ClusterAdmin) (sarama.SyncProducer, error) {
	log.Infof("开始初始化Kafka配置")
	defaultPartitionNum := config.GetConfig().Kafka.DefaultPartitionNum
	var defaultTopic = config.GetConfig().Topic.Topic
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = util.GenUUID()
	kafkaConfig.Version = sarama.MaxVersion
	// 使用 NewBalanceStrategyRange 函数来设置再平衡策略
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	kafkaConfig.Consumer.MaxProcessingTime = 2 * time.Second
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.RequiredAcks = sarama.NoResponse
	//kafkaConfig.Version = sarama.V2_8_1_0                          // Kafka版本根据实际情况选择
	//kafkaConfig.Net.TLS.Enable = true                              // 启用TLS
	kafkaConfig.Net.SASL.Enable = true                                // 启用SASL认证
	kafkaConfig.Net.SASL.User = config.GetConfig().Kafka.User         // Kafka认证用户名
	kafkaConfig.Net.SASL.Password = config.GetConfig().Kafka.Password // Kafka认证密码
	kafkaConfig.Net.SASL.Handshake = true                             // 启用SASL握手
	log.Infof("Kafka配置初始化完成，开始创建Producer")

	// 创建producer
	producer, err := sarama.NewSyncProducer([]string{config.GetConfig().Kafka.Addr}, kafkaConfig)
	if err != nil {
		log.Errorf("创建Producer失败: %v", err)
		return nil, err
	}
	log.Infof("Producer创建成功，开始初始化Admin客户端")

	// 初始化Kafka Admin客户端
	defer kafkaAdmin.Close()
	log.Infof("Admin客户端初始化成功，开始获取Topics列表")

	// 获取所有topics
	topics, err := kafkaAdmin.ListTopics()
	if err != nil {
		log.Errorf("获取Topics列表失败: %v", err)
		return nil, err
	}

	// 检查目标topic是否存在
	topicDetail, exists := topics[defaultTopic]
	if !exists {
		log.Infof("Topic[%s]不存在，开始创建新Topic", defaultTopic)
		// 如果topic不存在，创建新的topic
		topicDetail := sarama.TopicDetail{
			NumPartitions:     defaultPartitionNum,
			ReplicationFactor: 1,
		}
		err = kafkaAdmin.CreateTopic(defaultTopic, &topicDetail, false)
		if err != nil {
			log.Errorf("创建Topic[%s]失败: %v", defaultTopic, err)
			return nil, err
		}
		log.Infof("Topic[%s]创建成功", defaultTopic)
	} else if topicDetail.NumPartitions < defaultPartitionNum {
		log.Infof("Topic[%s]分区数[%d]小于配置分区数[%d]，开始更新分区", defaultTopic, topicDetail.NumPartitions, defaultPartitionNum)
		// 如果topic存在但分区数小于defaultPartitionNum，更新分区数
		err = kafkaAdmin.CreatePartitions(defaultTopic, defaultPartitionNum, nil, false)
		if err != nil {
			log.Errorf("更新Topic[%s]分区数失败: %v", defaultTopic, err)
			return nil, err
		}
		log.Infof("Topic[%s]分区数更新成功", defaultTopic)
	} else {
		log.Infof("Topic[%s]已存在且分区数[%d]符合要求", defaultTopic, topicDetail.NumPartitions)
	}

	log.Infof("Kafka初始化完成")
	return producer, nil
}

func sendMessageToKafka(msg interface{}, topic string) error {
	if msg == nil {
		return errors.New("message is nil")
	}
	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	kafkaMsg := &sarama.ProducerMessage{}
	kafkaMsg.Topic = topic
	kafkaMsg.Value = sarama.StringEncoder(message)
	log.Infof("kafka send topic: %s ;send msg>>>>>>>>>>>>> %s", topic, message)
	_, _, err = kafka.KafkaProducer.SendMessage(kafkaMsg)
	if err != nil {
		log.Errorf("kafka send topic: %s ;send error %v", topic, err)
		return err
	}
	return nil
}
