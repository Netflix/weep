## weep search account

Search for an account through ConsoleMe

### Synopsis

The search command allows users to search for resources via ConsoleMe. Currently, only
searching for accounts or roles is supported.



```
weep search account [query_string] [flags]
```

### Options

```
  -h, --help   help for account
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

* [weep search](weep_search.md)	 - Search for resources through ConsoleMe

