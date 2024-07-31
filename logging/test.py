import logging
import logging_loki
logging_loki.emitter.LokiEmitter.level_tag = "level"

handler = logging_loki.LokiHandler(
   url="http://127.0.0.1:27401/loki/api/v1/push",
   version="1",
)
logger = logging.getLogger("my-logger")
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

logger.error(
   "logging: Model3-initiated",
  
   extra={"tags": {"service": "my-service"}},
)
logger.warning(
   "logging: Model3-trained",
   extra={"tags": {"service": "my-service"}},
)

logger.info(
   "logging: Model4-trained",
   extra={"tags": {"service": "my-service"}},
)
logger.debug(
   "logging: Model4-deployed",
   extra={"tags": {"service": "my-service"}},
)