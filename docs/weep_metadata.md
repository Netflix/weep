## weep metadata

Run a local Instance Metadata Service (IMDS) endpoint that serves credentials

```
weep metadata [role_name] [flags]
```

### Options

```
  -h, --help                    help for metadata
  -a, --listen-address string   IP address for metadata service to listen on (default "127.0.0.1")
  -p, --port int                port for metadata service to listen on (default 9090)
  -r, --region string           region of metadata service (default "us-east-1")
```

### Options inherited from parent commands

```
  -c, --config string       config file (default is $HOME/.weep.yaml)
      --log-format string   log format (json or tty)
      --log-level string    log level (debug, info, warn)
```

### SEE ALSO

* [weep](weep.md)	 - weep helps you get the most out of ConsoleMe credentials

