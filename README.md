[![](https://img.shields.io/badge/docs-gitbook-blue`)](https://hawkins.gitbook.io/consoleme/weep-cli/)
[![Discord Chat](https://img.shields.io/discord/730908778299523072?label=Chat&logo=discord)](https://discord.gg/tZ8S7Yg)
[![pre-commit](https://github.com/Netflix/weep/actions/workflows/precommit.yml/badge.svg)](https://github.com/Netflix/weep/actions/workflows/precommit.yml)
[![goreleaser](https://github.com/Netflix/weep/actions/workflows/release.yml/badge.svg)](https://github.com/Netflix/weep/actions/workflows/release.yml)

# weep

Weep is a CLI utility for retreiving AWS credentials from [ConsoleMe](https://github.com/Netflix/consoleme). Weep can run
a local instance metadata service proxy, or export credentials as environment variables for your AWS needs. 

## Documentation

This README contains developer documentation. Weep user documentation can be found on [GitBook](https://hawkins.gitbook.io/consoleme/weep-cli/).

## Configuration

Weep can be compiled with an embedded configuration (See the Building section below), or it can get its configuration 
from a YAML-formatted file. We've included an example config file in [example-config.yaml](configs/example-config.yaml).

Weep searches for a configuration in the following locations:

- embedded configuration (see below)
- `/etc/weep/weep.yaml`
- `~/.weep/weep.yaml`
- `./weep.yaml`

Multiple configurations in these locations **will be merged** in the order listed above (e.g. entries in `./weep.yaml` will take precedence over `~/.weep/weep.yaml`.

You can also specify a config file as a CLI arg. This configuration will be used exclusively and will not be merged with other configurations:

```bash
weep --config somethingdifferent.yaml list
```

Weep supports authenticating to ConsoleMe in either a standalone challenge mode (ConsoleMe will authenticate the user
according to its settings), or mutual TLS (ConsoleMe has to be configured to accept mutual TLS).

In challenge mode, Weep will prompt the user for their username the first time they authenticate, and then attempt to
derive their username from their valid/expired jwt on subsequent attempts. You can also specify the desired username
in weep's configuration under the `challenge_settings.user` setting as seen in  `example-config.yaml`.

### Pre-Commit Setup
Weep uses pre-commit to run unit tests and Go linting.  Pre-commit documentation can be found on [pre-commit](https://pre-commit.com/)

#### Installation
You can install pre-commit using the following steps:

Using pip:
```
pip install pre-commit
```
Using [homebrew](https://brew.sh/):
```
brew install pre-commit
```
Using [Conda](https://conda.io/):
```
conda install -c conda-forge pre-commit
```

Validate your installation with the following:
```
$ pre-commit --version
pre-commit 2.9.3
```

#### Configuration
Set up the git hook scripts to run automatically with git commit
```
$ pre-commit install
pre-commit installed at .git/hooks/pre-commit
```

## Building

In most cases, `weep` can be built by running the `make` command in the repository root. `make release` (requires
[`upx`](https://upx.github.io/)) will build and compress the binary for distribution.

### Embedded configuration

`weep` binaries can be shipped with an embedded configuration to allow shipping an "all-in-one" binary.
An example of such a configuration is included in [example-config.yaml](example-config.yaml).

To compile with an embedded config, set the `EMBEDDED_CONFIG_FILE` environment variable at
build time. The value of this variable MUST be the **absolute path** of the configuration
file **relative to the root of the module**:

```bash
EMBEDDED_CONFIG_FILE=/example-config.yaml make
```

Note that the embedded configuration can be overridden by a configuration file in the locations listed above.

### Docker

#### Building and Running

```
make build-docker
docker run -v ~</optional/path/to/your/mtls/certs>:</optional/path/to/your/mtls/certs> --rm weep --meta-data --role <roleArn>
```

### Releasing

Weep uses [goreleaser](https://goreleaser.com/) in Github Actions for releases. Check their
[install docs](https://goreleaser.com/install/) if you would like to experiment with the release process locally.

To create a new release, create and push a tag using the release script (requires [svu](https://github.com/caarlos0/svu)):

```bash
./scripts/release.sh
```

Goreleaser will automatically create a release on the [Releases page](https://github.com/Netflix/weep/releases).

### Generating docs

Weep has a built-in command to generate command documentation (in the `docs/` directory):

```bash
weep docs
```
