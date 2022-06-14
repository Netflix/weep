## weep whoami

Print information about current AWS credentials

### Synopsis

The whoami command retrieves information about your AWS credentials from AWS STS using the default
credential provider chain. If SWAG (https://github.com/Netflix-Skunkworks/swag-api) is enabled, weep will
attempt to enrich the output with additional data.

```
weep whoami [flags]
```

### Options

```
  -h, --help   help for whoami
```

### Options inherited from parent commands

```
  -A, --assume-role strings        one or more roles to assume after retrieving credentials
  -c, --config string              config file (default is $HOME/.weep.yaml)
      --extra-config-file string   extra-config-file <yaml_file>
      --log-file string            log file path (default "/tmp/weep.log")
      --log-format string          log format (json or tty)
      --log-level string           log level (debug, info, warn)
  -n, --no-ip                      remove IP restrictions
  -r, --region string              AWS region (default "us-east-1")
```

### SEE ALSO

* [weep](weep.md)	 - weep helps you get the most out of ConsoleMe credentials

