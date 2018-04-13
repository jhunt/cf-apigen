package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jhunt/vcaptive"
)

func main() {
	url := os.Getenv("VAULT_URL")
	tok := os.Getenv("VAULT_TOKEN")
	pre := os.Getenv("VAULT_PREFIX")

	if os.Getenv("VCAP_SERVICES") != "" {
		services, err := vcaptive.ParseServices(os.Getenv("VCAP_SERVICES"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to parse VCAP_SERVICES: %s\n", err)
			os.Exit(1)
		}

		if vault, found := services.WithCredentials("vault"); found {
			if url, found = vault.GetString("vault"); !found {
				fmt.Fprintf(os.Stderr, "service %s does not define a vault URL (as `vault`)\n")
				os.Exit(1)
			}
			if tok, found = vault.GetString("token"); !found {
				fmt.Fprintf(os.Stderr, "service %s does not define a vault token (as `token`)\n")
				os.Exit(1)
			}
			if pre, found = vault.GetString("root"); !found {
				fmt.Fprintf(os.Stderr, "service %s does not define a vault prefix (as `root`)\n")
				os.Exit(1)
			}
		}
	}
	if url == "" {
		fmt.Fprintf(os.Stderr, "unable to determine Vault URL (did you forget to bind a Vault service?)\n")
		os.Exit(1)
	}
	if tok == "" {
		fmt.Fprintf(os.Stderr, "unable to determine Vault Token (did you forget to bind a Vault service?)\n")
		os.Exit(1)
	}
	if pre == "" {
		fmt.Fprintf(os.Stderr, "unable to determine Vault Prefix (did you forget to bind a Vault service?)\n")
		os.Exit(1)
	}

	port := ":8000"
	if os.Getenv("PORT") != "" {
		port = fmt.Sprintf(":%s", os.Getenv("PORT"))
	}

	api, err := Connect(url, tok, pre)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to use Vault at %s: %s\n", url, err)
		os.Exit(1)
	}

	http.Handle("/v1/", api)
	http.Handle("/", http.FileServer(http.Dir("htdocs")))
	fmt.Printf("vault cf-apigen app starting up...\n")
	http.ListenAndServe(port, nil)
}
