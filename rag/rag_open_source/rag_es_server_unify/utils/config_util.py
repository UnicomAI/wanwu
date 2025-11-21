from settings import ES_HOSTS, ES_USER, ES_PASSWORD, ES_VERIFY_CERTS
from elasticsearch import Elasticsearch

# Connect to Elasticsearch remote instance
es = Elasticsearch(
    hosts=ES_HOSTS,
    basic_auth=(ES_USER, ES_PASSWORD),
    verify_certs=ES_VERIFY_CERTS,  # Should be set to True in production environments to verify SSL certificates
    timeout=60,
    retry_on_timeout=True,
)

def get_health():
    # Get cluster health status, which includes node number information
    cluster_health = es.cluster.health()
    return cluster_health.body
