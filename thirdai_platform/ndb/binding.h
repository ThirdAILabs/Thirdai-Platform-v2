#pragma once

#include "stdbool.h"

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

void Document_add_metadata_bool(Document_t *doc, unsigned int i,
                                const char *key, bool value);
void Document_add_metadata_int(Document_t *doc, unsigned int i, const char *key,
                               int value);
void Document_add_metadata_float(Document_t *doc, unsigned int i,
                                 const char *key, float value);
void Document_add_metadata_str(Document_t *doc, unsigned int i, const char *key,
                               const char *value);

void NeuralDB_insert(NeuralDB_t *ndb, Document_t *doc, const char **err_ptr);

typedef struct MetadataList_t MetadataList_t;
void MetadataList_free(MetadataList_t *metadata);
unsigned int MetadataList_len(MetadataList_t *metadata);
const char *MetadataList_key(MetadataList_t *metadata, unsigned int i);
int MetadataList_type(MetadataList_t *metadata, unsigned int i);
bool MetadataList_bool(MetadataList_t *metadata, unsigned int i);
int MetadataList_int(MetadataList_t *metadata, unsigned int i);
float MetadataList_float(MetadataList_t *metadata, unsigned int i);
const char *MetadataList_str(MetadataList_t *metadata, unsigned int i);

typedef struct QueryResults_t QueryResults_t;
unsigned int QueryResults_len(QueryResults_t *results);
unsigned long long QueryResults_id(QueryResults_t *results, unsigned int i);
void QueryResults_free(QueryResults_t *results);
const char *QueryResults_text(QueryResults_t *results, unsigned int i);
const char *QueryResults_document(QueryResults_t *results, unsigned int i);
const char *QueryResults_doc_id(QueryResults_t *results, unsigned int i);
unsigned int QueryResults_doc_version(QueryResults_t *results, unsigned int i);
MetadataList_t *QueryResults_metadata(QueryResults_t *results, unsigned int i);
float QueryResults_score(QueryResults_t *results, unsigned int i);

QueryResults_t *NeuralDB_query(NeuralDB_t *ndb, const char *query,
                               unsigned int topk, const char **err_ptr);

void NeuralDB_save(NeuralDB_t *ndb, const char *save_path,
                   const char **err_ptr);

#ifdef __cplusplus
}
#endif
