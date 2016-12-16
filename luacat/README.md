# mcm-luacat

Build a catalog from a Lua script.

## Usage

```
mcm-luacat FILE
```

The `FILE` argument is a Lua script that is executed.
At the end of the script's execution, the catalog is written to stdout as binary Cap'n Proto data.

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
