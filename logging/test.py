import logging
import logging_loki
logging_loki.emitter.LokiEmitter.level_tag = "level"

handler = logging_loki.LokiHandler(
   url="http://192.168.1.11/loki/api/v1/push",
   version="1",
)
logger = logging.getLogger("my-logger")
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

# logger.info(
#    "test-1",
#    extra={"tags": {"deployment_id": str(123),"PER":21, "ORG": 51}},
# )

logger.info(
   "test-2",
   extra={"tags": {"deployment_id": 200,"PER":81, "ORG": 313}, "SSN": 5, "ADDR": 10},
)


logger.info(
   "test-2",
   extra={"tags": {"deployment_id": 200,"PER":81, "ORG": 313}, "SSN": 50, "ADDR": 100},
)

