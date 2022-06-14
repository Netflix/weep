## weep setup

Print setup information

### Synopsis

By default, this command will print a script that can be used to set up IMDS routing.
If you trust us enough, you can let Weep do all the work for you:

sudo weep setup --commit

Otherwise, run weep setup and inspect the output. Then save the output to a file or pass it to your shell:

# Pass to shell
weep setup  # trust no one, always inspect
weep setup | sudo sh

# Save to file
weep setup > setup.sh
cat setup.sh  # trust no one, always inspect
chmod u+x setup.sh
sudo ./setup.sh

```
weep setup [flags]
```

### Options

```
  -C, --commit   install all the things (probably requires root, definitely requires trust)
  -h, --help     help for setup
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

