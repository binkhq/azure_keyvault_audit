# Azure KeyVault Auditor

This app collects all of the keys in your Azure tenant and compares the production secrets values with the values of the secrets in your non-production vaults, to find production keys in these non-production vaults.

---

## Pre-requisites 
### Permissions 
#### Local
If you are running the app locally, you will need to have permission to read all of the vaults that are in scope of the application. 
You will then need to make sure you are authenticated using the Azure cli on your host. 
##### macOS
To install the Azure cli: `brew install azure-cli`
Once installed, run the following to authenticate interactively: `az login`
##### Windows & Linux 
Follow the relevant installation instructions here: [How to install the Azure CLI | Microsoft Docs](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
The same document going to into details on how to authenticate in each OS. 

---

## Setup
### Configuration 
The application can be configured via either a `config.toml` file or via environment variables. 
#### Configuration File 
The `config.toml` file contains the following settings: 
| Setting |Description  |Required? |Example |
|--|--|--|--|
|base_vault_url|The base of the url for accessing the vaults in Azure. A `%+v` format specifier is required.  | Required |`https://%+v.vault.azure.net`|
|prod_vaults |	A comma separated list of strings containing the vault names that are considered production within your tenant. | Required |`example-prod`|
|non_prod_vaults| A comma separated list of strings containing the vault names that are considered non-production within your tenant.|Required|`example-staging`|
|excluded_secrets| Allows you to specify secrets that are to be excluded from the audit. This will apply to all vaults.| Optional | `production-sftp-secret`|

The configuration file needs to be located in the same directory as the app binary. 
#### Environment Variables 
|Variable| Description|Required?|Example|
|--|--|--|--|
|`BASE_URL`| The base of the url for accessing the vaults in Azure. A `%+v` format specifier is required.| Required|`https://%+v.vault.azure.net`|
|`PROD_VAULTS` |	A comma separated list of strings containing the vault names that are considered production within your tenant. | Required |`"example1-prod,example2-prod"`|
|`NON_PROD_VAULTS`| A comma separated list of strings containing the vault names that are considered non-production within your tenant.|Required|`"example1-staging,example2-staging"`|
|`EXCLUDED_SECRETS`| Allows you to specify secrets that are to be excluded from the audit. This will apply to all vaults.| Optional | `"production-sftp-secret1,production-sftp-secret2"`|

---

### Installation / Execution - Local
#### Package 
Download on the prebuilt binaries from the release section for your OS. Place the configuration file in the same directory as the binary. 
##### Windows 
Open cmd in the current directory. 
Type `azure_keyvault_audit.exe` in the cli.
##### macOS & Linux 
Open a terminal in the current directory . 
Modify file permissions to be executable: `sudo chmod +x azure_keyvault_audit`
Run the application: `./azure_keyvault_audit`

#### From Source
Clone the repo into a local directory 
`git clone https://github.com/binkhq/azure_keyvault_audit.git`
Enter the directory
`cd azure_keyvault_audit`
Build the binary 
`go build`

#### Flags
|Flag|Type|Description|
|--|--|--|
|`-short`|bool| Provides a shortened output|
|`-debug`|bool| Turns on debug messages|