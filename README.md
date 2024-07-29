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
workos [cmd] [args]
```

### Environment Variables
WorkOS CLI support environment variables for initialization and environment management.

| Environment Variable              | Description                                                                                                                                   | Supported Values     |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|----------------------|
| WORKOS_ACTIVE_ENVIRONMENT         | Sets the selected environment in your .workos.json file. Use `env` to override environment configs with other environment variable overrides. |                      |
| WORKOS_ENVIRONMENTS_ENV_NAME      | Sets the name of the environment                                                                                                              |                      |
| WORKOS_ENVIRONMENTS_ENV_ENDPOINT  | Sets the base endpoint for the environment                                                                                                    |                      |
| WORKOS_ENVIRONMENTS_ENV_API_KEY   | Sets the API key for the environment                                                                                                          |                      |
| WORKOS_ENVIRONMENTS_ENV_TYPE      | Sets the env type for the environment                                                                                                         | Production / Sandbox |

#### Examples

##### Set the active environment

```shell
export WORKOS_ACTIVE_ENVIRONMENT=local
```

```json
// .workos.json
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
export WORKOS_ACTIVE_ENVIRONMENT=env
export WORKOS_ENVIRONMENTS_ENV_NAME=local
export WORKOS_ENVIRONMENTS_ENV_ENDPOINT=http://localhost:8001
export WORKOS_ENVIRONMENTS_ENV_API_KEY=<YOUR_KEY>
export WORKOS_ENVIRONMENTS_ENV_TYPE=Sandbox
```
