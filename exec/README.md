# mcm-exec

Apply a catalog to the local system.

## Usage

```
mcm-exec [-n] [-q] [-s] [CATALOG]
```

If the CATALOG argument is omitted, then it is read from stdin.
`-n` activates dry-run mode: any potentially system-changing operations do nothing and report success.
`-q` suppresses normal informative output.
`-s` shows underlying operations as they occur.
