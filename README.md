# mcm &mdash; Minimal Configuration Manager

<b>Minimal Configuration Manager (mcm)</b> is a suite of tools to provide configuration management, like Puppet or Chef.
However, mcm differs in that it tries to embody the [Unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy), in particular: <q>Make each program do one thing well. To do a new job, build afresh rather than complicate old programs by adding new "features".</q>
mcm operates on a data format for describing a catalog of resources, with tools for both producing and consuming this data format.

**This is not an official Google product; it is just code that happens to be owned by Google.**

## Basic Usage

```bash
$ cat > foo.lua <<EOF
-- mcm operates on catalogs, which you can create using Lua.
-- The exact schema is documented in catalog.capnp.

mcm.resource('hello', {}, mcm.file{
  path = "/etc/hello.txt",
  plain = {content = "Hello, World!\n"},
})
EOF

$ mcm-luacat foo.lua > foo.out     # Run the Lua script and create a
                                   # catalog file.
$ sudo mcm-exec foo.out            # mcm-exec applies a catalog directly
                                   # to the local system.
$ mcm-shellify < foo.out > foo.sh  # mcm-shellify converts a catalog
                                   # into a self-contained shell script.
                                   # The script can be used to run over
                                   # SSH, etc.
```

## Building

mcm is built using [Bazel](https://bazel.build/).
A script is provided to automatically download and install Bazel for you.
The only dependency you must install is a modern C++ compiler.
Bazel will take care of fetching the rest of the dependencies.

```bash
$ cd path/to/mcm
$ ./bazel build //...  # Build all binaries
$ ./bazel test //...   # Run all tests
```

## Discuss

Join the [mcm-develop](https://groups.google.com/forum/#!forum/mcm-develop) group.

Please follow the [Go Code of Conduct](https://golang.org/conduct) when posting in both the group and the issue tracker.
If you encounter any conduct-related issues, contact Ross Light (ross@zombiezen.com).

## License

Apache 2.0.  See the [LICENSE](LICENSE) file for details.

[Contributing?](CONTRIBUTING.md)
