---
title: anyquery tool mysql-password
description: Learn how to use the anyquery tool mysql-password command in Anyquery.
---

Hash a password from stdin to be used in an authentification file

### Synopsis

Hash a password from stdin to be used in an authentification file.
The password is hashed using the mysql_native_password algorithm
which can be summarized as HEX(SHA1(SHA1(password)))

```bash
anyquery tool mysql-password [flags]
```

### Examples

```bash
echo "password" | anyquery tool mysql-password
```

### Options

```bash
  -h, --help   help for mysql-password
```

### SEE ALSO

* [anyquery tool](../anyquery_tool)	 - Tools to help you with using anyquery
