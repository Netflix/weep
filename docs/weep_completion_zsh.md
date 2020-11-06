## weep completion zsh

Generate Zsh completion

### Synopsis

To load completions:

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for each session, execute once:

$ weep completion zsh > "${fpath[1]}/_weep"

You will need to start a new shell for this setup to take effect.


```
weep completion zsh [flags]
```

### Options

```
  -h, --help   help for zsh
```

### Options inherited from parent commands

```
  -c, --config string       config file (default is $HOME/.weep.yaml)
      --log-format string   log format (json or tty)
      --log-level string    log level (debug, info, warn)
```

### SEE ALSO

* [weep completion](weep_completion.md)	 - Generate completion script

