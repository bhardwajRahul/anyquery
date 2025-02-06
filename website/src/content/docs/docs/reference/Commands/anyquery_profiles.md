---
title: anyquery profiles
description: Learn how to use the anyquery profiles command in Anyquery.
---

Print the profiles installed on the system

### Synopsis

Print the profiles installed on the system.
Alias to profile list.

```bash
anyquery profiles [registry] [plugin] [flags]
```

### Examples

```bash
# List the profiles
anyquery profiles
```

### Options

```bash
  -c, --config string   Path to the configuration database
      --csv             Output format as CSV
      --format string   Output format (pretty, json, csv, plain)
  -h, --help            help for profiles
      --json            Output format as JSON
      --plain           Output format as plain text
```

### SEE ALSO

* [anyquery](../anyquery)	 - A tool to query any data source
* [anyquery profiles delete](../anyquery_profiles_delete)	 - Delete a profile
* [anyquery profiles list](../anyquery_profiles_list)	 - List the profiles
* [anyquery profiles new](../anyquery_profiles_new)	 - Create a new profile
* [anyquery profiles update](../anyquery_profiles_update)	 - Update the profiles configuration
