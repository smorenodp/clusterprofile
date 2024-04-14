# ClusterProfile

## Description

ClusterProfile is designed to allow users to authenticate against all the services of a cluster using Vault as the primary entity. Once the user is authenticated against Vault, ClusterProfile generates the credentials and information needed to use all of those services generating a bash executable file to export all the required variables in the shell.

## Installation

You only have to compile the go binary with go build.

## Usage

To use ClusterProfile you first need to create the configuration file that contains all the information to load for the profile selected, this configuration file can be divided in multiple ones, for example, creating a file for each profile. By default the folder is in $HOME/.clusterid/profiles/ but can be overrided with the environment variable CLUSTERID_CONFIG_FOLDER or with the option profileFolder in the binary.

The content of the file looks something like this:

```yaml
- name: test
  vault:
    addr: localhost:8200
    method: oidc
  providers:
  - type: nomad
    addr: localhost:4646
    backend: nomad
    method: role
    config:
      role: self
  - type: consul
    addr: localhost:8500
    backend: consul
    method: role
    config:
      role: self
```

* name - name of the profile, later used when exporting it with the binary
* vault - configuration for the vault server, you need to specify the address to connect to and the method to login (ATM: oidc or token)
* providers - configuration for the services deployed in the cluster with the info required to authenticate against each one with vault.
  * type - service type (ATM: consul or nomad)
  * addr - address of the service in case the provider needs it
  * backend - name of the backend in vault (by default it's the type)
  * method - login method (ATM: role or token)
  * config - all the config needed for the login method, in case of the role we only need the role name (Available config: role, token, policies)

Once you create this file you can execute clusterprofile to load the creds for a certain profile:

```bash
clusterprofile -profile test
```

The binary options are:

* profile - To indicate which profile to generate/load credentials
* profileFolder - change the default folder for the profile configuration (By default in $HOME/.clusteid/profiles/)
* creds - change the default file for the profile credentials (By default in $HOME/.clusteid/credentials)
* exec - change the default file for the executable files (By default in $HOME/.clusteid/export.sh)

Once you execute the binary, it checks if the credentials were already generated and are not expired in the credentials file in wich the values are stored each time the binary is executed:

```text
[test]
VAULT_TOKEN="***"
VAULT_TTL="2024-04-15 09:59:40"
VAULT_ADDR="localhost:8200"
NOMAD_TOKEN="***"
NOMAD_TTL="2024-04-15 09:59:40"
NOMAD_ADDR="localhost:4646"
CONSUL_HTTP_TOKEN="***"
CONSUL_TTL="2024-04-15 09:59:40"
CONSUL_HTTP_ADDR="localhost:8500"                                                     
```

If we load the profile again later but this values are still usable, it won't create new values and load this instead.

To load this variables in the shell, the binary creates an export file containing all this information:

```bash

#!/bin/bash
echo "Loading test credentials"

export VAULT_TOKEN="***"

export VAULT_TTL="2024-04-15 09:59:40"

export VAULT_ADDR="localhost:8200"

export NOMAD_TOKEN="***"

export NOMAD_TTL="2024-04-15 09:59:40"

export NOMAD_ADDR="localhost:4646"

export CONSUL_HTTP_TOKEN="***"

export CONSUL_TTL="2024-04-15 09:59:40"

export CONSUL_HTTP_ADDR="localhost:8500"
```

Which can then be exported with `source $HOME/.clusteid/export.sh`

## Pivot profile

An advanced configuration for a profile would be to use a pivoting profile, with this you can configure anothe profile to load before generating the credentials for the target profile. This is needed in case that, prior to generate the target profile credentials, you need access to the services/vault with certain permissions.

```yaml
- name: test-pivot
  vault:
    addr: localhost:8200
    pivoting_profile: test
    method: token
    config:
      role: admin
      policies:
      - admin
  providers:
  - type: nomad
    addr: localhost:4646
    backend: nomad
    method: role
    config:
      role: admin
  - type: consul
    addr: localhost:8500
    backend: consul
    method: role
    config:
      role: admin
```

In this case, when loading the profile **test-pivot**, clusterprofile would first load the credentials for the profile test and, using this information, generate the ones for the target profile, which are:

* Creating a Vault token with role *admin* and policies *admin*
* Creating a Nomad token with role admin in the backend (nomad/creds/admin)
* Creating a Consul token with role admin in the backend (consul/creds/admin)
