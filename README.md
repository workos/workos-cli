# WorkOS CLI

## Installation

### MacOS (Homebrew)

```shell
brew install workos/tap/workos-cli
```

## Authentication

To authenticate the WorkOS CLI with the WorkOS API, use a new or existing API key from the WorkOS dashboard.

```shell
workos init
```

Follow the prompts after running the `init` command to configure the CLI to use the specified API key.

The CLI can be configured to use multiple API keys to make it easy to manage data in multiple environments.

```shell
workos apikey add
```

To switch between API keys, use the `apikey switch` command and select the API key you would like to use:

```shell
workos apikey switch
```

To remove a configured API key from the CLI, use the `apikey remove` command and select the API key you would like to remove:

```shell
workos apikey remove
```
