## weep search

Search for resources through ConsoleMe

### Synopsis

The search command allows users to search for resources via ConsoleMe. Currently, only
searching for accounts or roles is supported.



### Options

```
  -h, --help   help for search
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
* [weep search account](weep_search_account.md)	 - Search for an account through ConsoleMe
* [weep search role](weep_search_role.md)	 - Search for a role in an account through ConsoleMe

