from thirdai import neural_db_v2 as ndb

db = ndb.NeuralDB()

db.insert([ndb.PDF("/home/david/du_faqs.pdf", version="v1")])

# results = db.search("how to pay my bill", 5)
# query = "how to pay my bill"
query = "what is the maximum refund when I unsubscribe"
# query = "where to get cash refunds"
results = db.search(query, 5)

for chunk, score in results:
    print(chunk.text, "\n")

context = "\n\n".join(chunk.text for chunk, _ in results)



# PHI PROMPTING
# import requests
# import json

# url = "http://34.236.171.80:80/on-prem-llm/completion"
# headers = {"Content-Type": "application/json"}
# data = {
    # "system_prompt": "Answer the user's questions based on the given context.",
    # "prompt": "<|user|>" + context + f"\n given this context, {query}?<|end|>\n<|assistant|>",
#     # "stream": True,
#     "n_predict": 1000,
# }

# response = requests.post(url, headers=headers, json=data)

# print(json.loads(response.text)["content"])


# with requests.post(url, headers=headers, json=data, stream=True) as response:
#     for chunk in response.iter_content(chunk_size=8192):
#         if chunk:
#             print(chunk.decode("utf-8"))







# # LLAMA PROMPTING
# import requests
# import json

# url = "http://34.236.171.80:80/on-prem-llm/completion"
# headers = {"Content-Type": "application/json"}
# data = {
#     "system_prompt": "",
#     "prompt": f"""
#     <|begin_of_text|>
#     <|start_header_id|>system<|end_header_id|>
#     Answer the user's questions based on the given context.<|eot_id|>

#     Context: {context}<|eot_id|>

#     <|start_header_id|>user<|end_header_id|>
#     Given this context, {query}?<|eot_id|>

#     <|start_header_id|>assistant<|end_header_id|>
#     """,
#     # "stream": True,
#     "n_predict": 1000,
# }

# response = requests.post(url, headers=headers, json=data)

# print(json.loads(response.text)["content"])






# # LLAMA PROMPTING
# import requests
# import json

# url = "http://34.236.171.80:80/on-prem-llm/completion"
# headers = {"Content-Type": "application/json"}
# data = {
#     "system_prompt": "",
#     "prompt": f"""
#     <|begin_of_text|>
#     <|start_header_id|>system<|end_header_id|>
#     Answer the user's questions based on the given context.<|eot_id|>

#     <|start_header_id|>user<|end_header_id|>
#     {context} \n Given this context, {query}?<|eot_id|>

#     <|start_header_id|>assistant<|end_header_id|>
#     """,
#     # "stream": True,
#     "n_predict": 1000,
# }

# response = requests.post(url, headers=headers, json=data)

# print(json.loads(response.text)["content"])