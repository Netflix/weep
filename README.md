# weep

Weep is a CLI utility for retreiving AWS credentials from ConsoleMe. Weep can run
a local instance metadata service proxy, or export credentials as environmental
variables for your AWS needs. 


## Configuration

Make a weep configuration file in one of the following locations:

- `./.weep.yaml`
- `~/.weep.yaml`
- `~/.config/weep/.weep.yaml`

You can also specify a config file as a CLI arg:

```
weep --config somethingdifferent.yaml list
```

## Routing traffic

### Mac

```bash
sudo ifconfig lo0 169.254.169.254 alias
echo "rdr pass on lo0 inet proto tcp from any to 169.254.169.254 port 80 -> 127.0.0.1 port 9090" | sudo pfctl -ef -
```

#### Persisting Changes
Plist files are located in [extras/com.user.lo0-loopback.plist](extras/com.user.lo0-loopback.plist) and [extras/com.user.weep.plist](extras/com.user.weep.plist)

To persist the settings above on a Mac, download the plists and place them in `/Library/LaunchDaemons` and
reboot or issue the following commands:

```bash
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

### Metadata Proxy

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

## Building

In most cases, `weep` can be built by running the `make` command in the repository root. `make release` (requires
[`upx`](https://upx.github.io/)) will build and compress the binary for distribution.

### Embedding mTLS configuration

`weep` binaries can be shipped with an embedded mutual TLS (mTLS) configuration to 
avoid making users set this configuration. An example of such a configuration is included
in [mtls/mtls_paths.yaml](mtls/mtls_paths.yaml).

To compile with an embedded config, set the `MTLS_CONFIG_FILE` environment variable at
build time. The value of this variable MUST be the **absolute path** of the configuration
file **relative to the root of the module**:

```bash
MTLS_CONFIG_FILE=/mtls/mtls_paths.yaml make
```

### Docker

#### Building and Running

```
make build-docker
docker run -v ~</optional/path/to/your/mtls/certs>:</optional/path/to/your/mtls/certs> --rm weep --meta-data --role <roleArn>
```

#### Publishing a Docker image

To publish a Docker image, you can invoke `make docker`, which runs `make build-docker` and `make publish-docker`. When run from any branch other than `master`, the image is tagged with the version number and branch name. On the `master` branch the image is tagged with only the version number.

> To update the version number, change the `VERSION` variable in `Makefile`.
