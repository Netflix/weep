# weep

Weep is a CLI utility for retreiving AWS credentials from ConsoleMe. Weep can run
a local instance metadata service proxy, or export credentials as environmental
variables for your AWS needs. 


## Configuration

Weep can be compiled with an embedded configuration (See the Building section below), or it can get its configuration 
from a YAML-formatted file. We've included an example config file in [example-config.yaml](example-config.yaml).

Weep searches for a configuration file in the following locations:

- `./.weep.yaml`
- `~/.weep.yaml`
- `~/.config/weep/.weep.yaml`

You can also specify a config file as a CLI arg:

```
weep --config somethingdifferent.yaml list
```

Weep supports authenticating to ConsoleMe in either a standalone challenge mode (ConsoleMe will authenticate the user
according to its settings), or mutual TLS (ConsoleMe has to be configured to accept mutual TLS).

In challenge mode, Weep will prompt the user for their username the first time they authenticate, and then attempt to
derive their username from their valid/expired jwt on subsequent attempts. You can also specify the desired username
in weep's configuration under the `challenge_settings.user` setting as seen in  `example-config.yaml`.


## Routing traffic

### Mac

```bash
sudo ifconfig lo0 169.254.169.254 alias
echo "rdr pass on lo0 inet proto tcp from any to 169.254.169.254 port 80 -> 127.0.0.1 port 9090" | sudo pfctl -ef -
```

#### Persisting Changes

You can look at the recommended plist files in [extras/com.user.lo0-loopback.plist](extras/com.user.lo0-loopback.plist) and [extras/com.user.weep.plist](extras/com.user.weep.plist)

To persist the settings above on a Mac, download the plists, place them in `/Library/LaunchDaemons`, and load them
using `launchctl`:

> **Note:** Make sure you know what you're doing here -- these commands change system behavior.

```bash
curl https://raw.githubusercontent.com/Netflix/weep/master/extras/com.user.weep.plist -o com.user.weep.plist
curl https://raw.githubusercontent.com/Netflix/weep/master/extras/com.user.lo0-loopback.plist -o com.user.lo0-loopback.plist
sudo mv com.user.weep.plist com.user.lo0-loopback.plist /Library/LaunchDaemons/
launchctl load /Library/LaunchDaemons/com.user.weep.plist
launchctl load /Library/LaunchDaemons/com.user.lo0-loopback.plist
```


### Linux

```bash
# trap all output packets to metadata proxy and send them to localhost:9090
iptables -t nat -A OUTPUT -p tcp --dport 80 -d 169.254.169.254 -j DNAT --to 127.0.0.1:9090
```

To persist this, create a txt file at the location of your choosing with the 
following contents:

```
*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [1:216]
:POSTROUTING ACCEPT [1:216]
-A OUTPUT -d 169.254.169.254/32 -p tcp -m tcp --dport 80 -j DNAT --to-destination 127.0.0.1:9090
COMMIT
```

Enable the rules by running the following:

sudo /sbin/iptables-restore < <path_to_file>.txt

## Usage

### ECS Credential Provider (Recommended)

Weep supports emulating the ECS credential provider to provide credentials to your AWS SDK. This solution can be
minimally configured by setting the `AWS_CONTAINER_CREDENTIALS_FULL_URI` environment variable for your process. There's
no need for iptables or routing rules with this approach, and each different shell or process can use weep to request
credentials for different roles. Weep will cache the credentials you request in-memory, and will refresh them on-demand
when they are within 10 minutes of expiring.

In one shell, run weep:

```bash
weep ecs_credential_provider
```

In your favorite IDE or shell, set the `AWS_CONTAINER_CREDENTIALS_FULL_URI` environment variable and run AWS commands.

```bash
AWS_CONTAINER_CREDENTIALS_FULL_URI=http://localhost:9091/ecs/consoleme_oss_1 aws sts get-caller-identity
{
    "UserId": "AROA4JEFLERSKVPFT4INI:user@example.com",
    "Account": "123456789012",
    "Arn": "arn:aws:sts::123456789012:assumed-role/consoleme_oss_1_test_user/user@example.com"
}

AWS_CONTAINER_CREDENTIALS_FULL_URI=http://localhost:9091/ecs/consoleme_oss_2 aws sts get-caller-identity
{
    "UserId": "AROA6KW3MOV2F7J6AT4PC:user@example.com",
    "Account": "223456789012",
    "Arn": "arn:aws:sts::223456789012:assumed-role/consoleme_oss_2_test_user/user@example.com"
}
```

### Metadata Proxy

Weep supports emulating the instance metadata service. This requires that you have iptables DNAT rules configured, and
it only serves one role per weep process.

```bash
# You can use a full ARN
weep metadata arn:aws:iam::123456789012:role/exampleRole

# ...or just the role name
weep metadata exampleRole
```

run `aws sts get-caller-identity` to confirm that your DNAT rules are correctly configured.

### Credential export

```bash
eval $(weep export arn:aws:iam::123456789012:role/fullOrPartialRoleName)

# this one also works with just the role name!
eval $(weep export fullOrPartialRoleName)
```

Then run `aws sts get-caller-identity` to confirm that your credentials work properly.

### Credentials file

Write retrieved credentials to an AWS credentials file (`~/.aws/credentials` by default with the profile name `consoleme`).

```bash
weep file exampleRole

# you can also specify a profile name
weep file stagingRole --profile staging
weep file prodRole --profile prod

# or you can save it to a different place
weep file exampleRole -o /tmp/credentials
```

Weep will do its best to preserve existing credentials in the file (but it will overwrite a conflicting profile name, so be careful!).

### Credentials Process
The AWS CLI can source credentials from weep using the `credential_process` configuration which can be defined for a
profile in the `~/.aws/config` file. Read more about this process [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html).

Here's an example of an `~/.aws/config` file:

```bash
[profile role1]
credential_process = /path/to/weep credential_process role1

[profile role2]
credential_process = /path/to/weep credential_process role2
```

To use the credential process, you would invoke the AWS CLI with the `AWS_PROFILE` environment variable set  to the
profile you wanted to use. Example:

```bash
AWS_PROFILE=role1 aws s3 ls
```

#### Generating Credential Process Commands

Weep can also generate credential process commands and populate your ~/.aws/config file. 

**CAUTION**

AWS SDKs appear to be analyzing your ~/.aws/config file on each API call
and this could drastically slow you down if your ~/.aws/config file is too large. We strongly recommend using Weep's 
ECS credential provider to avoid this issue.

```bash
# Please read the caveat above before running this command. The size of your ~/.aws/config file may negatively impact 
# the rate of your AWS API calls.
weep generate_credential_process_config
```
## Shell Completion

### Bash

```bash
source <(weep completion bash)
```

To load completions for each session, execute this command once:

```bash
# Linux:
weep completion bash > /etc/bash_completion.d/weep
# MacOS:
weep completion bash > /usr/local/etc/bash_completion.d/weep
```

### Zsh
If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

To load completions for each session, execute this command once:

```bash
weep completion zsh > "${fpath[1]}/_weep"
```

You will need to start a new shell for this setup to take effect.

### Fish

```bash
weep completion fish | source
```

To load completions for each session, execute this command once:

```bash
weep completion fish > ~/.config/fish/completions/weep.fish
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
