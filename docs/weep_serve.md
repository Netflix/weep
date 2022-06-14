## weep serve

Run a local ECS Credential Provider endpoint that serves and caches credentials for roles on demand

### Synopsis

The serve command runs a local webserver that serves the /ecs/ path. When the
AWS_CONTAINER_CREDENTIALS_FULL_URI environment variable is set to a URL, the 
AWS CLI and SDKs will use that URL to retrieve credentials. For example, if 
you want to use credentials for a role called SuperCoolRole, you could do 
something like this:

AWS_CONTAINER_CREDENTIALS_FULL_URI=http://localhost:9091/ecs/SuperCoolRole \
        aws sts get-caller-identity

If you just want to use a single role, use the 'role' positional argument to specify which one and it
will be served the same way credentials are served in an EC2 instance. There’s no need
to set an environment variable for this.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-provider


```
weep serve [optional_role_name] [flags]
```

### Options

```
  -h, --help                    help for serve
  -a, --listen-address string   IP address for the ECS credential provider to listen on (default "127.0.0.1")
  -p, --port int                port for the ECS credential provider service to listen on (default 9091)
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

