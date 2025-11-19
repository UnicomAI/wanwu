
from log.log_config import setup_logging
from settings import APP_NAME, LOGGER_NAME

# Set up logs and get loggers
logger = setup_logging(APP_NAME, LOGGER_NAME)