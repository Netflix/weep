## weep open

Generate (and open) a ConsoleMe link for a given ARN

### Synopsis

The open command generates the link for supported resources in ConsoleMe. By default, this command 
also attempts to open the browser after generating the link. Use the --no-open flag to prevent opening. 
The supported resources match those that are supported by ConsoleMe. IAM roles, s3, sqs and sns resources open in the ConsoleMe editor, while other supported resources attempt to redirect to the AWS Console using the right role.


```
weep open <arn> [flags]
```

### Options

```
  -h, --help      help for open
  -x, --no-open   don't automatically open links
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

