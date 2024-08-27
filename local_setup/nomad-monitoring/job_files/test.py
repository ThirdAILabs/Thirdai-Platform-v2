import logging

import logging_loki

logging_loki.emitter.LokiEmitter.level_tag = "level"

handler = logging_loki.LokiHandler(
    url="http://192.168.1.6/loki/api/v1/push", version="1"
)
logger = logging.getLogger("my-logger")
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

logger.error(
    "test-1", extra={"tags": {"group": "victoria-loki", "PER": 21, "ORG": 51}}
)

# logger.info(
#    "test-2",
#    extra={"tags": {"group": "loki","PER":81, "ORG": 313}}
# )


# logger.info(
#    "test-2",
#    extra={"tags": {"deployment_id": 200,"PER":81, "ORG": 313}, "SSN": 50, "ADDR": 100},
# )


# import json
# import requests
# import time
# # Loki endpoint URL and headers
# url = "http://192.168.1.6:81/loki/api/v1/push"
# headers = {
#     'Content-type': 'application/json'
# }

# # Message to log
# msg = "This is the message"

# # Construct the payload
# payload = {
#     'streams': [
#         {
#             'stream': {
#                 'group': 'loki',
#                 'level': 'info'
#             },
#             'values': [
#                 [str(int(time.time() * 1e9)), msg]  # Timestamp in nanoseconds, log line
#             ]
#         }
#     ]
# }

# # Convert payload to JSON
# payload_json = json.dumps(payload)

# # Send the request
# response = requests.post(url, data=payload_json, headers=headers)

# # Print response
# print(response.status_code)
# print(response.text)
