---
title: anyquery completion zsh
description: Learn how to use the anyquery completion zsh command in Anyquery.
---

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(anyquery completion zsh)

To load completions for every new session, execute once:

#### Linux:

	anyquery completion zsh > "${fpath[1]}/_anyquery"

#### macOS:

	anyquery completion zsh > $(brew --prefix)/share/zsh/site-functions/_anyquery

You will need to start a new shell for this setup to take effect.


```bash
anyquery completion zsh [flags]
```

### Options

```bash
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [anyquery completion](../anyquery_completion)	 - Generate the autocompletion script for the specified shell
