package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	secrets          []secret_obj
	prod_secrets     []vault
	non_prod_secrets []vault
	base_url         string
	non_prod_vaults  []string
	prod_vaults      []string
	excluded_secrets []string
)

type secret_obj struct {
	name  string
	value string
}

type vault struct {
	name    string
	secrets []secret_obj
}

func get_secrets(vault string, cred *azidentity.DefaultAzureCredential, excluded_secrets []string) {

	client, err := azsecrets.NewClient(vault, cred, nil)
	if err != nil {
		log.Error().Msgf("Unable to create client for the following vault: %+v", err)
	}
	pager := client.ListSecrets(nil)

	for pager.NextPage(context.TODO()) {
		pager_resp := pager.PageResponse()

		for _, secret := range pager_resp.Secrets {
			secret_ID := strings.SplitAfterN(*secret.ID, "secrets/", 2)
			if exclusion_check(secret_ID[1], excluded_secrets) {
				continue
			}
			secret_value := get_secret_value(client, secret_ID[1])
			secrets = append(secrets, secret_obj{secret_ID[1], secret_value})
		}
	}
	if pager.Err() != nil {
		log.Error().Msgf("Client Pager error: %+v", pager.Err())
	}
}

func exclusion_check(secret string, excluded []string) bool {
	for _, s := range excluded {
		if secret == s {
			return true
		}
	}
	return false
}

func get_secret_value(client *azsecrets.Client, secret_ID string) string {

	secret_resp, err := client.GetSecret(context.TODO(), secret_ID, nil)
	if err != nil {
		log.Error().Msgf("unable to access %+v value: %+v", secret_ID, err)
	}
	secret_value := *secret_resp.Value
	return secret_value
}

func secret_find(short *bool) {
	log.Debug().Msg("Checking for prod secrets")
	fmt.Println(*short)
	for _, v := range prod_secrets {
		for _, ps := range v.secrets {

			for _, t := range non_prod_secrets {
				for _, nps := range t.secrets {
					if ps.value == nps.value {
						if *short {
							log.Warn().
								Str("Production_Secret_Name", ps.name).
								Str("Production_Vault_Name", v.name).
								Str("Non_Production_Secret_Name", nps.name).
								Str("Non_Production_Vault_Name", t.name).
								Msg("Match")
						} else {
							log.Warn().
								Str("Production_Secret_Name", ps.name).
								Str("Production_Vault_Name", v.name).
								Str("Non_Production_Secret_Name", nps.name).
								Str("Non_Production_Vault_Name", t.name).
								Msgf("The secret %v in %v matches the secret %v in %v", ps.name, v.name, nps.name, t.name)
						}

					}
				}
			}
		}
	}
}
func setting_check() {

	if _, err := os.Stat("config.toml"); err == nil {
		log.Debug().Msg("Using config file for settings")
		config, _ := toml.LoadFile("config.toml")
		base_url = config.Get("base_vault_url").(string)
		non_prod_vaults = config.GetArray("non_prod_vaults").([]string)
		prod_vaults = config.GetArray("prod_vaults").([]string)
		excluded_secrets = config.GetArray("excluded_secrets").([]string)
	} else if errors.Is(err, os.ErrNotExist) {
		log.Debug().Msg("Using enviroment variables for settings")
		base_url = os.Getenv("BASE_URL")
		non_prod_vaults = strings.Split(os.Getenv("NON_PROD_VAULTS"), ",")
		prod_vaults = strings.Split(os.Getenv("PROD_VAULTS"), ",")
		excluded_secrets = strings.Split(os.Getenv("EXCLUDED_SECRETS"), ",")
	} else {
		log.Fatal().Msg("No config supplied, exiting")
	}
}

func main() {

	debug := flag.Bool("debug", false, "Sets log level to debug")
	short := flag.Bool("short", false, "Outputs a shorter JSON log")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msg("Debug Logging enabled")

	log.Info().Msg("Starting KeyVault Audit")

	setting_check()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Authentication failure")
	}

	log.Debug().Msg("Getting Non-prod secrets")
	for _, v := range non_prod_vaults {
		vault_url := fmt.Sprintf(base_url, v)
		log.Debug().Msgf("Getting secrets in %v", v)
		get_secrets(vault_url, cred, excluded_secrets)
		log.Debug().Msgf("%v complete", v)
		var kv = vault{v, secrets}
		non_prod_secrets = append(non_prod_secrets, kv)
		secrets = nil
	}

	log.Debug().Msg("Getting prod secrets")
	for _, v := range prod_vaults {
		vault_url := fmt.Sprintf(base_url, v)
		log.Debug().Msgf("Getting secrets in %v", v)
		get_secrets(vault_url, cred, excluded_secrets)
		log.Debug().Msgf("%v complete", v)
		var kv = vault{v, secrets}
		prod_secrets = append(prod_secrets, kv)
		secrets = nil
	}

	secret_find(short)
	non_prod_secrets = nil
	prod_secrets = nil

}
