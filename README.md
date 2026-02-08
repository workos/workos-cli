# WorkOS CLI

## Installation

### MacOS (Homebrew)

#### Install

```shell
brew install workos/tap/workos-cli
```

#### Upgrade

```shell
brew upgrade workos/tap/workos-cli
```

## Usage

First, initialize the CLI:

```shell
workos init
```

Follow the interactive prompts to configure the CLI for use with the specified environment.

The CLI can be configured to work with multiple WorkOS environments.

```shell
workos env add
```

To switch between environments, use the `env switch` command and select the environment you would like to switch to:

```shell
workos env switch
```

To remove a configured environment from the CLI, use the `env remove` command and select the environment you would like to remove:

```shell
workos env remove
```

Once initialized, the CLI is ready to use:

```shell
workos
The WorkOS CLI is a tool to interact with WorkOS APIs via the command line.

Usage:
  workos [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  env          Manage configured environments
  fga          Manage FGA resources (resource types, warrants, and resources).
  help         Help about any command
  init         Initialize the CLI
  organization Manage organizations (create, update, delete, etc).
  user         Manage users (get, list, update, delete, etc).

Flags:
  -h, --help               help for workos
      --timeout duration   Timeout for commands (default 10s)
  -v, --version            version for workos

Use "workos [command] --help" for more information about a command.
```

### Environment Variables

WorkOS CLI support environment variables for initialization and environment management.

| Environment Variable                  | Description                                                                                                                                        | Supported Values     |
|---------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|----------------------|
| WORKOS_ACTIVE_ENVIRONMENT             | Sets the selected environment in your .workos.json file. Use `headless` to override environment configs with other environment variable overrides. |                      |
| WORKOS_ENVIRONMENTS_HEADLESS_NAME     | Sets the name of the environment                                                                                                                   |                      |
| WORKOS_ENVIRONMENTS_HEADLESS_ENDPOINT | Sets the base endpoint for the environment                                                                                                         |                      |
| WORKOS_ENVIRONMENTS_HEADLESS_API_KEY  | Sets the API key for the environment                                                                                                               |                      |
| WORKOS_ENVIRONMENTS_HEADLESS_TYPE     | Sets the env type for the environment                                                                                                              | Production / Sandbox |

#### Examples

##### Set the active environment

```shell
export WORKOS_ACTIVE_ENVIRONMENT=local
```

`.workos.json`

```json
{
  "environments": {
    "local": {
      "endpoint": "http://localhost:8001",
      "apiKey": "<YOUR_KEY>",
      "type": "Sandbox",
      "name": "local"
    }
  }
}
```

##### Headless Mode

```shell
export WORKOS_ACTIVE_ENVIRONMENT=headless
export WORKOS_ENVIRONMENTS_HEADLESS_NAME=local
export WORKOS_ENVIRONMENTS_HEADLESS_ENDPOINT=http://localhost:8001
export WORKOS_ENVIRONMENTS_HEADLESS_API_KEY=<YOUR_KEY>
export WORKOS_ENVIRONMENTS_HEADLESS_TYPE=Sandbox
```
