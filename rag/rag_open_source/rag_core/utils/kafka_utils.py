from kafka import KafkaProducer
import os
import json
import time
import datetime
from logging_config import setup_logging
logger_name='rag_kafka_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))
from settings import FLYWHEEL_KAFKA_TOPIC, KAFKA_BOOTSTRAP_SERVERS


def create_kafka_producer(bootstrap_servers):
    producer = KafkaProducer(
        bootstrap_servers=bootstrap_servers,
        value_serializer=lambda x: json.dumps(x).encode('utf-8')
    )
    return producer


def send_message(producer, topic, message):
    try:
        producer.send(topic, value=message)
        producer.flush()
        logger.info(f"Sent message to {topic}: {message}")
    except Exception as e:
        logger.error(f"Failed to send message to {topic}: {e}")
        # raise


def push_kafka_msg(message):
    logger.info("=======>push_kafka_msg")
    print("=======>push_kafka_msg")
    bootstrap_servers = KAFKA_BOOTSTRAP_SERVERS
    topic = FLYWHEEL_KAFKA_TOPIC

    # 创建 Kafka 生产者 [EN] Create a Kafka producer
    producer = create_kafka_producer(bootstrap_servers)

    
    send_message(producer, topic, message)

    # 关闭生产者 [EN] shutdown producer
    producer.close()


if __name__ == "__main__":
    # main()

    doc = {
        "user_id": "222",
        "kb_name": "b",
        "question": "hello world",
        "rate": 0.4,
        "top_k": 5,
        "default_answer": "unicom-70b",
        "model_name": "model_name",
        "search_field": "emc",
        "search_list": [],
        "status": 0,
        "created": int(round(time.time() * 1000)),
        "dt": int(datetime.datetime.now().strftime("%Y%m%d"))
    }
    push_kafka_msg(doc)