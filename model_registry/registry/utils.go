package registry

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

func getSecret() []byte {
	// This is only used for jwt secrets, if the server restarts the only issue is any
	// tokens issued before the restart (that aren't yet expired) will be invalidated.
	b := make([]byte, 16)

	n, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	if n != len(b) {
		panic(err)
	}

	return b
}

func requestParsingError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("Error parsing request body: %v", err), http.StatusBadRequest)
}

func dbError(w http.ResponseWriter, err error) {
	if err == gorm.ErrRecordNotFound {
		http.Error(w, fmt.Sprintf("Unable to retrieve record for request: %v", err), http.StatusBadRequest)
	} else {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
	}
}
