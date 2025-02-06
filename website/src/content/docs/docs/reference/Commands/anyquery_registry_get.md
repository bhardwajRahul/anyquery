---
title: anyquery registry get
description: Learn how to use the anyquery registry get command in Anyquery.
---

Print informations about a registry

```bash
anyquery registry get [name] [flags]
```

### Examples

```bash
anyquery registry get internal_reg
```

### Options

```bash
      --csv             Output format as CSV
      --format string   Output format (pretty, json, csv, plain)
  -h, --help            help for get
      --json            Output format as JSON
      --plain           Output format as plain text
```

### Options inherited from parent commands

```bash
  -c, --config string   Path to the configuration database
```

### SEE ALSO

* [anyquery registry](../anyquery_registry)	 - List the registries where plugins can be downloaded
