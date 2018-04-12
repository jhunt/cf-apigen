package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/starkandwayne/safe/vault"
	"gopkg.in/yaml.v2"
)

var Config struct {
	VaultURL   string `yaml:"vault_url"`
	VaultToken string `yaml:"vault_token"`
}

var VaultServer *vault.Vault

func CreateToken(w http.ResponseWriter, r *http.Request) {
	var token struct {
		Token string `json:"token"`
	}
	var email struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&email)
	tokenPath := fmt.Sprintf("secret/tokens/%s", email.Email)

	newSecret := vault.NewSecret()
	newSecret.Password("token", 16, "a-f0-9", false)

	err := VaultServer.Write(tokenPath, newSecret)
	if err != nil {
		log.Printf("Error:%s\n", err.Error())
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
		return
	}

	apiToken, err := VaultServer.Read(tokenPath)
	if err != nil {
		log.Printf("Error:%s\n", err.Error())
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
		return
	}

	token.Token = apiToken.Get("token")
	json.NewEncoder(w).Encode(token)
}

func main() {
	err := ReadConfig("cf-apigen.conf")
	if err != nil {
		log.Printf("Error:%s\n", err.Error())
	}
	log.Printf("Targeting Vault at %s with token %s\n", Config.VaultURL, Config.VaultToken)
	VaultServer, err = vault.NewVault(Config.VaultURL, Config.VaultToken, true)
	if err != nil {
		log.Printf("Error:%s\n", err.Error())
	}

	port := ":8000"
	if os.Getenv("PORT") != "" {
		port = fmt.Sprintf(":%s", os.Getenv("PORT"))
	}

	router := mux.NewRouter()
	router.HandleFunc("/v1/token", CreateToken).Methods("POST")
	log.Fatal(http.ListenAndServe(port, router))
}

func ReadConfig(file string) error {
	if file != "" {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if err = yaml.Unmarshal(b, &Config); err != nil {
			return err
		}
	}
	return nil
}
