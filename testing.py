from thirdai import neural_db as ndb

db = ndb.NeuralDB()

db.insert([ndb.PDF("/home/david/du_faqs.pdf", version="v1")])

results = db.search("how to pay my bill", 5)

context = ""

for chunk, score in results:
    print(chunk.text, "\n")


# Questions
# ndbv1 or v2 for chat and search
# pdf parser v1 or v2
# whats the prompt
# whats the context radius