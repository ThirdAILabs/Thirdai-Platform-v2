#include "binding.h"
#include "include/OnDiskNeuralDB.h"
#include <algorithm>
#include <cstring>
#include <iostream>
#include <memory>
#include <optional>
#include <vector>

using thirdai::search::ndb::Chunk;
using thirdai::search::ndb::MetadataMap;
using thirdai::search::ndb::OnDiskNeuralDB;

void copyError(const std::exception &e, const char **err_ptr) {
  char *err_msg = new char[std::strlen(e.what())];
  std::strcpy(err_msg, e.what());
  *err_ptr = err_msg;
}

struct NeuralDB_t {
  std::unique_ptr<OnDiskNeuralDB> ndb;

  NeuralDB_t(const std::string &save_path)
      : ndb(OnDiskNeuralDB::make(save_path)) {}
};

NeuralDB_t *NeuralDB_new(const char *save_path, const char **err_ptr) {
  try {
    std::string path(save_path);
    return new NeuralDB_t(path);
  } catch (const std::exception &e) {
    // TODO(Nicholas): have case for NeuralDBError to return better errors
    copyError(e, err_ptr);
    return nullptr;
  }
}

void NeuralDB_free(NeuralDB_t *ndb) { delete ndb; }

struct Document_t {
  std::vector<std::string> chunks;
  // std::vector<MetadataMap> metadata;
  std::string document;
  std::string doc_id;
  std::optional<uint32_t> doc_version;
};

Document_t *Document_new(const char *document, const char *doc_id) {
  Document_t *doc = new Document_t();
  doc->document = document;
  doc->doc_id = doc_id;

  return doc;
}

void Document_free(Document_t *doc) { delete doc; }

void Document_add_chunk(Document_t *doc, const char *chunk) {
  doc->chunks.emplace_back(chunk);
}

void Document_set_version(Document_t *doc, unsigned int version) {
  doc->doc_version = version;
}

void NeuralDB_insert(NeuralDB_t *ndb, Document_t *doc, const char **err_ptr) {
  try {
    ndb->ndb->insert(
        /*chunks=*/doc->chunks,
        /*metadata*/ std::vector<MetadataMap>(doc->chunks.size()),
        /*document=*/doc->document,
        /*doc_id=*/doc->doc_id,
        /*doc_version=*/doc->doc_version);
  } catch (const std::exception &e) {
    copyError(e, err_ptr);
    return;
  }
}

struct QueryResults_t {
  std::vector<std::pair<Chunk, float>> results;
};

void QueryResults_free(QueryResults_t *results) { delete results; }

unsigned int QueryResults_len(QueryResults_t *results) {
  return results->results.size();
}

unsigned long long QueryResults_id(QueryResults_t *results, unsigned int i) {
  return results->results.at(i).first.id;
}

const char *QueryResults_text(QueryResults_t *results, unsigned int i) {
  return results->results.at(i).first.text.c_str();
}

const char *QueryResults_document(QueryResults_t *results, unsigned int i) {
  return results->results.at(i).first.document.c_str();
}

const char *QueryResults_doc_id(QueryResults_t *results, unsigned int i) {
  return results->results.at(i).first.doc_id.c_str();
}

unsigned int QueryResults_doc_version(QueryResults_t *results, unsigned int i) {
  return results->results.at(i).first.doc_version;
}

float QueryResults_score(QueryResults_t *results, unsigned int i) {
  return results->results.at(i).second;
}

QueryResults_t *NeuralDB_query(NeuralDB_t *ndb, const char *query,
                               unsigned int topk, const char **err_ptr) {
  try {
    auto results = ndb->ndb->query(query, topk);
    auto out = new QueryResults_t();
    out->results = results;
    return out;
  } catch (const std::exception &e) {
    // TODO(Nicholas): have case for NeuralDBError to return better errors
    copyError(e, err_ptr);
    return nullptr;
  }
}

void NeuralDB_save(NeuralDB_t *ndb, const char *save_path,
                   const char **err_ptr) {
  try {
    ndb->ndb->save(save_path);
  } catch (const std::exception &e) {
    copyError(e, err_ptr);
    return;
  }
}