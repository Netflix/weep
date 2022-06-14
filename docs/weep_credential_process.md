## weep credential_process

Retrieve credentials on the fly via the AWS SDK

### Synopsis

The credential_process command can be used by AWS SDKs to retrieve 
credentials from Weep on the fly. The --generate flag lets you automatically
generate an AWS configuration with profiles for all of your available roles, or 
you can manually update your configuration (see the link below to learn how).

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-process


```
weep credential_process [role_name] [flags]
```

### Options

```
  -g, --generate        generate ~/.aws/config with credential process config
  -h, --help            help for credential_process
  -o, --output string   output file for AWS config (default "/Users/jdhulia/.aws/config")
  -p, --pretty          when combined with --generate/-g, use 'account_name-role_name' format for generated profiles instead of arn
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

