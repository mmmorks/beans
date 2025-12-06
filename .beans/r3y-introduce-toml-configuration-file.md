---
title: Introduce TOML configuration file
status: done
created_at: 2025-12-06T14:05:22.233408Z
updated_at: 2025-12-06T14:34:38Z
---



Add a TOML configuration file to allow customization of beans behavior.

## Requirements

- Configuration file location: `.beans/beans.toml`
- The `beans init` command should create this file with sensible defaults
- Use [pelletier/go-toml](https://github.com/pelletier/go-toml) for parsing

## Initial Configuration Options

For now, the configuration will define available statuses:

```toml
[statuses]
available = ["open", "in-progress", "done"]
default = "open"
```

## Future Extensions

This configuration file will be extended to support additional settings as the project evolves.