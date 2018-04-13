package main

import (
	"os"
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/starkandwayne/safe/vault"
)

type API struct {
	Vault *vault.Vault
	Prefix string
}

func Connect(url, token, prefix string) (API, error) {
	vault, err := vault.NewVault(url, token, true)
	return API{
		Vault: vault,
		Prefix: prefix,
	}, err
}

func (the API) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/v1/token" {
		if req.Method == "POST" {
			var in struct {
				Email string `json:"email"`
			}
			json.NewDecoder(req.Body).Decode(&in)
			tokenPath := fmt.Sprintf("%s/tokens/%s", the.Prefix, in.Email)

			tok := vault.NewSecret()
			tok.Password("token", 16, "a-f0-9", false)

			err := the.Vault.Write(tokenPath, tok)
			if err != nil {
				bail(w, err)
				return
			}

			tok, err = the.Vault.Read(tokenPath)
			if err != nil {
				bail(w, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(
				struct {
					Token string `json:"token"`
				}{
					Token: tok.Get("token"),
				})
			return
		}

		w.WriteHeader(405)
		fmt.Fprintf(w, "method %s not allowed (must be POST)\n", req.Method)
		return
	}
	w.WriteHeader(404)
	fmt.Fprintf(w, "%s: not a cf-apigen endpoint\n", req.URL.Path)
}

func bail(w http.ResponseWriter, err error) {
	fmt.Fprintf(os.Stderr, "oops: %s\n", err)
	w.WriteHeader(500)
	fmt.Fprintf(w, "%s\n", err)
}
