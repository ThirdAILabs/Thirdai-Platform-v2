import logging
import logging_loki
logging_loki.emitter.LokiEmitter.level_tag = "level"

handler = logging_loki.LokiHandler(
   url="http://localhost:3100/loki/api/v1/push",
   version="1",
)
logger = logging.getLogger("my-logger")
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

logger.error(
   "logging: Model1-initiated",
  
   extra={"tags": {"service": "my-service"}},
)
logger.warning(
   "logging: Model1-trained",
   extra={"tags": {"service": "my-service"}},
)

logger.info(
   "logging: Model2-trained",
   extra={"tags": {"service": "my-service"}},
)
logger.debug(
   "logging: Model2-deployed",
   extra={"tags": {"service": "my-service"}},
)