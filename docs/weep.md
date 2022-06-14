## weep

weep helps you get the most out of ConsoleMe credentials

### Synopsis

Weep is a CLI tool that manages AWS access via ConsoleMe for local development.

### Options

```
  -A, --assume-role strings        one or more roles to assume after retrieving credentials
  -c, --config string              config file (default is $HOME/.weep.yaml)
      --extra-config-file string   extra-config-file <yaml_file>
  -h, --help                       help for weep
      --log-file string            log file path (default "/tmp/weep.log")
      --log-format string          log format (json or tty)
      --log-level string           log level (debug, info, warn)
  -n, --no-ip                      remove IP restrictions
  -r, --region string              AWS region (default "us-east-1")
```

### SEE ALSO

* [weep console](weep_console.md)	 - Log into the AWS Management console
* [weep credential_process](weep_credential_process.md)	 - Retrieve credentials on the fly via the AWS SDK
* [weep export](weep_export.md)	 - Retrieve credentials to be exported as environment variables
* [weep file](weep_file.md)	 - Retrieve credentials and save them to a credentials file
* [weep list](weep_list.md)	 - List available roles
* [weep open](weep_open.md)	 - Generate (and open) a ConsoleMe link for a given ARN
* [weep search](weep_search.md)	 - Search for resources through ConsoleMe
* [weep serve](weep_serve.md)	 - Run a local ECS Credential Provider endpoint that serves and caches credentials for roles on demand
* [weep setup](weep_setup.md)	 - Print setup information
* [weep version](weep_version.md)	 - Print version information
* [weep whoami](weep_whoami.md)	 - Print information about current AWS credentials

