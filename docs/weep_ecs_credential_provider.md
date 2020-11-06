## weep ecs_credential_provider

Run a local ECS Credential Provider endpoint that serves and caches credentials for roles on demand

```
weep ecs_credential_provider [flags]
```

### Options

```
  -h, --help                    help for ecs_credential_provider
  -a, --listen-address string   IP address for the ECS credential provider to listen on (default "127.0.0.1")
  -p, --port int                port for the ECS credential provider service to listen on (default 9090)
```

### Options inherited from parent commands

```
  -c, --config string       config file (default is $HOME/.weep.yaml)
      --log-format string   log format (json or tty)
      --log-level string    log level (debug, info, warn)
```

### SEE ALSO

* [weep](weep.md)	 - weep helps you get the most out of ConsoleMe credentials

