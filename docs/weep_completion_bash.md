## weep completion bash

Generate Bash completion

### Synopsis

To load completions:

$ source <(weep completion bash)

To load completions for each session, execute once:

Linux:

  $ weep completion bash > /etc/bash_completion.d/weep

MacOS:

  $ weep completion bash > /usr/local/etc/bash_completion.d/weep


```
weep completion bash [flags]
```

### Options

```
  -h, --help   help for bash
```

### Options inherited from parent commands

```
  -c, --config string       config file (default is $HOME/.weep.yaml)
      --log-format string   log format (json or tty)
      --log-level string    log level (debug, info, warn)
```

### SEE ALSO

* [weep completion](weep_completion.md)	 - Generate completion script

