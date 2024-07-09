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
