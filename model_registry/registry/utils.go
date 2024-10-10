package registry

import "crypto/rand"

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
