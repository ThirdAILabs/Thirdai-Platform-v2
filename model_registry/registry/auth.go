package registry

import "github.com/go-chi/jwtauth/v5"

var tokenAuth *jwtauth.JWTAuth

func Init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}
