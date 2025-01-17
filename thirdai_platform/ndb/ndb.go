package ndb

// #cgo LDFLAGS: -L. -lthirdai -lrocksdb -lutf8proc -L/opt/homebrew/opt/libomp/lib/ -lomp
// #cgo CXXFLAGS: -fPIC -std=c++17 -I. -fvisibility=hidden
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

func (ndb *NeuralDB) Insert(document, doc_id string, chunks []string, version *uint) error {
	doc := newDocument(document, doc_id)
	defer C.Document_free(doc)
	for _, chunk := range chunks {
		addChunk(doc, chunk)
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
	// Metadata   map[string]interface{}
	Score float32
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
	}

	return chunks, nil
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
