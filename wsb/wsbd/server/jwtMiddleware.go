package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

const pubKeyPath = "config/jwt/public.pem"

var verifyKey *rsa.PublicKey

func init() {
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["token"]
		if !ok || len(keys) < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "Url Param 'token' is missing!")

			return
		}

		jwtToken, err := jwt.Parse(keys[0], func(token *jwt.Token) (interface{}, error) { return verifyKey, nil })
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "Invalid Token!")

			return
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok && !jwtToken.Valid {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		ctx := context.WithValue(
			r.Context(),
			sessionKey,
			Session{Identifier: claims["username"].(string)},
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}