package main

import (
	"credentials"
	"document"
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/definition/", handlerGetWordByID).Methods("GET")
	muxRouter.HandleFunc("/definition", handlerGetWord).Methods("GET")
	muxRouter.HandleFunc("/definition", handlerPostWord).Methods("POST")
	muxRouter.HandleFunc("/definition", handlerPutWord).Methods("PUT")
	muxRouter.HandleFunc("/definition", handlerDeleteWord).Methods("DELETE")
	muxRouter.HandleFunc("/", handlerSignIn)
	muxRouter.HandleFunc("/welcome", handlerWelcome)
	return muxRouter
}

func handlerGetWord(w http.ResponseWriter, r *http.Request) {
	movies, err := dictionary.FindAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, movies)
}

func handlerGetWordByID(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	word, err := dictionary.FindByValue(query.Get("word"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, word)
}

func handlerPostWord(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var word document.Word
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&word); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	word.ID = bson.NewObjectId()
	err := dictionary.Insert(word)
	switch {
	case mgo.IsDup(err):
		respondWithError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, word)
}

func handlerPutWord(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Not implemented yet putdef")
}

func handlerDeleteWord(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Not implemented yet putdef")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		http.Error(w, "HTTP 500: Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(response)
}

//Create JSON web token
func createToken(w http.ResponseWriter, r *http.Request) {
	//Decode client response into User struct
	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	//Create JWT with signing method and claims(i.e. payload)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": user.Username,
			"password": user.Password,
		},
	)

	//Sign JWT with secret signingKey
	signingKey := "signJwtUsingSecretKey"
	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//Respond with the token
	respondWithJSON(w, http.StatusCreated, credentials.JWToken{Token: tokenString})
}

func verifyToken(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	jwt.Parse(query["token"][0])

	word, err := dictionary.FindByValue(query.Get("word"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, word)
}
