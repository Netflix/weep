## weep list

List available roles

### Synopsis

The list command prints out all of the roles you have access to via ConsoleMe. By default,
this command will only show console roles. Use the --all flag to also include application
roles.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/list-eligible-roles


```
weep list [flags]
```

### Options

```
  -a, --account string   filter by aws account number or account name
      --all              show all available roles (default option) (default true)
  -e, --extended-info    include additional information about roles such as associated apps
  -h, --help             help for list
  -i, --instance         show only instance roles
  -p, --profiles         show only configured roles
  -s, --short-info       only display the role ARNs
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

