package ndb

// #cgo linux LDFLAGS: -L. -lthirdai -lrocksdb -lutf8proc -fopenmp
// #cgo darwin LDFLAGS: -L. -lthirdai -lrocksdb -lutf8proc -L/opt/homebrew/opt/libomp/lib/ -lomp
// #cgo CXXFLAGS: -O3 -fPIC -std=c++17 -I. -fvisibility=hidden
// #include "binding.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

type NeuralDB struct {
	ndb *C.NeuralDB_t
}

func New(savePath string) (NeuralDB, error) {
	savePathCStr := C.CString(savePath)
	defer C.free(unsafe.Pointer(savePathCStr))

	var err *C.char
	ndb := C.NeuralDB_new(savePathCStr, &err)
	if err != nil {
		defer C.free(unsafe.Pointer(err))
		return NeuralDB{}, errors.New(C.GoString(err))
	}

	return NeuralDB{ndb: ndb}, nil
}

func (ndb *NeuralDB) Free() {
	C.NeuralDB_free(ndb.ndb)
}

func newDocument(document, doc_id string) *C.Document_t {
	documentCStr := C.CString(document)
	defer C.free(unsafe.Pointer(documentCStr))
	docIdCStr := C.CString(doc_id)
	defer C.free(unsafe.Pointer(docIdCStr))

	doc := C.Document_new(documentCStr, docIdCStr)

	return doc
}

func addChunk(doc *C.Document_t, chunk string) {
	chunkCStr := C.CString(chunk)
	defer C.free(unsafe.Pointer(chunkCStr))
	C.Document_add_chunk(doc, chunkCStr)
}

func addMetadata(doc *C.Document_t, i int, key string, value interface{}) {
	keyCStr := C.CString(key)
	defer C.free(unsafe.Pointer(keyCStr))

	switch value := value.(type) {
	case bool:
		C.Document_add_metadata_bool(doc, C.uint(i), keyCStr, C.bool(value))
	case int:
		C.Document_add_metadata_int(doc, C.uint(i), keyCStr, C.int(value))
	case float32:
		C.Document_add_metadata_float(doc, C.uint(i), keyCStr, C.float(value))
	case float64:
		C.Document_add_metadata_float(doc, C.uint(i), keyCStr, C.float(value))
	case string:
		valueCStr := C.CString(value)
		defer C.free(unsafe.Pointer(valueCStr))
		C.Document_add_metadata_str(doc, C.uint(i), keyCStr, valueCStr)
	}
}

func (ndb *NeuralDB) Insert(document, doc_id string, chunks []string, metadata []map[string]interface{}, version *uint) error {
	doc := newDocument(document, doc_id)
	defer C.Document_free(doc)
	for _, chunk := range chunks {
		addChunk(doc, chunk)
	}

	if metadata != nil {
		for i, m := range metadata {
			for k, v := range m {
				addMetadata(doc, i, k, v)
			}
		}
	}

	if version != nil {
		C.Document_set_version(doc, C.uint(*version))
	}

	var err *C.char
	C.NeuralDB_insert(ndb.ndb, doc, &err)
	if err != nil {
		defer C.free(unsafe.Pointer(err))
		return errors.New(C.GoString(err))
	}

	return nil
}

type Chunk struct {
	Id         uint64
	Text       string
	Document   string
	DocId      string
	DocVersion uint32
	Metadata   map[string]interface{}
	Score      float32
}

func (ndb *NeuralDB) Query(query string, topk int) ([]Chunk, error) {
	queryCStr := C.CString(query)
	defer C.free(unsafe.Pointer(queryCStr))

	var err *C.char
	results := C.NeuralDB_query(ndb.ndb, queryCStr, C.uint(topk), &err)
	if err != nil {
		defer C.free(unsafe.Pointer(err))
		return nil, errors.New(C.GoString(err))
	}
	defer C.QueryResults_free(results)

	nResults := C.uint(C.QueryResults_len(results))
	chunks := make([]Chunk, nResults)
	for i := C.uint(0); i < nResults; i++ {
		chunks[i].Id = uint64(C.QueryResults_id(results, i))
		chunks[i].Text = C.GoString(C.QueryResults_text(results, i))
		chunks[i].Document = C.GoString(C.QueryResults_document(results, i))
		chunks[i].DocId = C.GoString(C.QueryResults_doc_id(results, i))
		chunks[i].DocVersion = uint32(C.QueryResults_doc_version(results, i))
		chunks[i].Score = float32(C.QueryResults_score(results, i))
		chunks[i].Metadata = convertMetadata(C.QueryResults_metadata(results, i))
	}

	return chunks, nil
}

func convertMetadata(metadata *C.MetadataList_t) map[string]interface{} {
	defer C.MetadataList_free(metadata)

	len := C.MetadataList_len(metadata)
	out := make(map[string]interface{})

	for i := C.uint(0); i < len; i++ {
		key := C.GoString(C.MetadataList_key(metadata, i))
		switch C.MetadataList_type(metadata, i) {
		case 0:
			out[key] = bool(C.MetadataList_bool(metadata, i))
		case 1:
			out[key] = int(C.MetadataList_int(metadata, i))
		case 2:
			out[key] = float32(C.MetadataList_float(metadata, i))
		case 3:
			out[key] = C.GoString(C.MetadataList_str(metadata, i))
		}
	}
	return out
}

func (ndb *NeuralDB) Save(savePath string) error {
	savePathCStr := C.CString(savePath)
	defer C.free(unsafe.Pointer(savePathCStr))

	var err *C.char
	C.NeuralDB_save(ndb.ndb, savePathCStr, &err)
	if err != nil {
		defer C.free(unsafe.Pointer(err))
		return errors.New(C.GoString(err))
	}

	return nil
}
