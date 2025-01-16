#pragma once

#ifdef __cplusplus
extern "C" {
#endif

typedef struct NeuralDB_t NeuralDB_t;

NeuralDB_t *NeuralDB_new(const char *save_path, const char **err_ptr);

void NeuralDB_free(NeuralDB_t *ndb);

typedef struct Document_t Document_t;
Document_t *Document_new(const char *document, const char *doc_id);
void Document_free(Document_t *doc);
void Document_add_chunk(Document_t *doc, const char *chunk);
void Document_set_version(Document_t *doc, unsigned int version);

void NeuralDB_insert(NeuralDB_t *ndb, Document_t *doc, const char **err_ptr);

typedef struct QueryResults_t QueryResults_t;
unsigned int QueryResults_len(QueryResults_t *results);
unsigned long long QueryResults_id(QueryResults_t *results, unsigned int i);
void QueryResults_free(QueryResults_t *results);
const char *QueryResults_text(QueryResults_t *results, unsigned int i);
const char *QueryResults_document(QueryResults_t *results, unsigned int i);
const char *QueryResults_doc_id(QueryResults_t *results, unsigned int i);
unsigned int QueryResults_doc_version(QueryResults_t *results, unsigned int i);
float QueryResults_score(QueryResults_t *results, unsigned int i);

QueryResults_t *NeuralDB_query(NeuralDB_t *ndb, const char *query,
                               unsigned int topk, const char **err_ptr);

void NeuralDB_save(NeuralDB_t *ndb, const char *save_path,
                   const char **err_ptr);

#ifdef __cplusplus
}
#endif
