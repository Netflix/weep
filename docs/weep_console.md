## weep console

Log into the AWS Management console

### Synopsis

The login command opens a browser window with a link that will log you into the
AWS Management console using the specified role. You can use the --no-open flag to simply print the console
link, rather than opening it in a browser.


```
weep console [flags]
```

### Options

```
  -h, --help      help for console
  -x, --no-open   print the link, but do not open a browser window
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

