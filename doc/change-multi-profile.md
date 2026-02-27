# Multi-Server Profile Support

## Config Format

```yaml
# ~/.gerrit/.config.yml (new format)
default: work

profiles:
  work:
    server: http://172.21.45.228
    username: fog
    password: "123456"
    project: STB_SCC
    no_auth_prefix: true
  oss:
    server: https://gerrit-review.googlesource.com
    username: fog
    password: "some-token"
```

Usage:
```
gerrit change list                    # uses "default" profile (work)
gerrit --profile oss change list      # uses oss profile
GERRIT_PROFILE=oss gerrit change list # env var alternative
```

## Design

Flatten the active profile's values into viper via `viper.Set()` in `initConfig()`. This means `api/client.go` and all command files need zero changes — they keep reading `viper.GetString("server")` etc.

## Backward Compatibility

- Old flat config detected by `LoadMulti()` (no `profiles:` key → legacy) and wrapped as single "default" profile in memory
- File on disk stays flat until user runs `gerrit init` again, which auto-migrates to new format
- `--config` flag bypasses profile logic entirely (existing behavior)

## Changes

| File | Change |
|------|--------|
| `internal/config/config.go` | Rename `Config` → `Profile` (keep alias). Add `MultiConfig` struct, `LoadMulti()`, `SaveMulti()`, `ResolveProfile()` |
| `internal/cmd/root/root.go` | Add `--profile` flag. Rewrite `initConfig()` to load profile and flatten into viper |
| `internal/cmd/init/init.go` | Add profile name prompt. Use `LoadMulti`/`SaveMulti` to merge profiles |
| `api/client.go` | No changes |
| All command files | No changes |

## Status: Implemented
