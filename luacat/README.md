# mcm-luacat

Build a catalog from a Lua script.

## Usage

```
mcm-luacat [-o FILE] [-I PATTERN [...]] SCRIPT
```

The `SCRIPT` argument is the path to a Lua script that is executed.
At the end of the script's execution, the catalog is written to stdout (or to the file named by the `-o` flag) as binary Cap'n Proto data.

### `require` Search Path

The script's containing directory is added to `package.path`, specifically as `DIR/?.lua;DIR/?/init.lua`.
The `-I` flag and the `MCM_LUACAT_PATH` environment variable can be used to add paths to `package.path` beyond the script's directory.
These arguments are interpreted as in [`package.searchpath`](https://www.lua.org/manual/5.3/manual.html#pdf-package.searchpath) &mdash; semicolon-separated templates containing `?` wildcards.
The search order is:

1.  Script's directory
2.  Any include paths added via the `-I` flag
3.  Any include paths added via the `MCM_LUACAT_PATH` environment variable

## The `mcm` package

The Lua script environment will have an `mcm` package loaded in the globals table.

```lua
mcm.resource(id, deps, resource)
```

The primary function in the package.
`id` can be a string or an id (as returned by `mcm.hash`).
`deps` is a table list of other resource IDs -- again, either strings or ids.
`resource` is a table as returned by one of the resource type functions below.

```lua
mcm.file(table)
mcm.exec(table)
mcm.noop
```

The return value of these functions (or `mcm.noop`) are used as the third argument to `mcm.resource`.
Each one of the functions takes in a table whose fields correspond with the struct inside [catalog.capnp](../catalog.capnp).
`mcm.noop` is a value for the no-op resource type.

```lua
mcm.hash(s)
```

Returns an id based on the content of a string.
Useful for referencing ids in resource types, like `Exec.condition.ifDepsChanged`.
