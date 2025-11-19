package mq

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"

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
	kafkaConfig.Net.SASL.Enable = true                                // 启用SASL认证 [EN] Enable SASL authentication
	kafkaConfig.Net.SASL.User = config.GetConfig().Kafka.User         // Kafka认证用户名 [EN] Kafka authentication username
	kafkaConfig.Net.SASL.Password = config.GetConfig().Kafka.Password // Kafka认证密码 [EN] Kafka authentication password
	kafkaConfig.Net.SASL.Handshake = true                             // 启用SASL握手 [EN] Enable SASL handshake
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
	var defaultTopic = config.GetConfig().Kafka.Topic
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = util.UUID()
	kafkaConfig.Version = sarama.MaxVersion
	// 使用 NewBalanceStrategyRange 函数来设置再平衡策略 [EN] Use the NewBalanceStrategyRange function to set the rebalancing strategy
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	kafkaConfig.Consumer.MaxProcessingTime = 2 * time.Second
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.RequiredAcks = sarama.NoResponse
	//kafkaConfig.Version = sarama.V2_8_1_0                          // Kafka版本根据实际情况选择 [EN] kafkaConfig.Version = sarama.V2_8_1_0 //The Kafka version is selected according to the actual situation
	//kafkaConfig.Net.TLS.Enable = true                              // 启用TLS [EN] kafkaConfig.Net.TLS.Enable = true // Enable TLS
	kafkaConfig.Net.SASL.Enable = true                                // 启用SASL认证 [EN] Enable SASL authentication
	kafkaConfig.Net.SASL.User = config.GetConfig().Kafka.User         // Kafka认证用户名 [EN] Kafka authentication username
	kafkaConfig.Net.SASL.Password = config.GetConfig().Kafka.Password // Kafka认证密码 [EN] Kafka authentication password
	kafkaConfig.Net.SASL.Handshake = true                             // 启用SASL握手 [EN] Enable SASL handshake
	log.Infof("Kafka配置初始化完成，开始创建Producer")

	// 创建producer [EN] Create producer
	producer, err := sarama.NewSyncProducer([]string{config.GetConfig().Kafka.Addr}, kafkaConfig)
	if err != nil {
		log.Errorf("创建Producer失败: %v", err)
		return nil, err
	}
	log.Infof("Producer创建成功，开始初始化Admin客户端")

	// 初始化Kafka Admin客户端 [EN] Initialize the Kafka Admin client
	defer kafkaAdmin.Close()
	log.Infof("Admin客户端初始化成功，开始获取Topics列表")

	// 获取所有topics [EN] Get all topics
	topics, err := kafkaAdmin.ListTopics()
	if err != nil {
		log.Errorf("获取Topics列表失败: %v", err)
		return nil, err
	}

	// 检查目标topic是否存在 [EN] Check whether the target topic exists
	topicDetail, exists := topics[defaultTopic]
	if !exists {
		log.Infof("Topic[%s]不存在，开始创建新Topic", defaultTopic)
		// 如果topic不存在，创建新的topic [EN] If the topic does not exist, create a new topic
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
		// 如果topic存在但分区数小于defaultPartitionNum，更新分区数 [EN] If the topic exists but the number of partitions is less than defaultPartitionNum, update the number of partitions
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

func SendMessage(msg interface{}, topic string) error {
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
