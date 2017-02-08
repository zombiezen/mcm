---
layout: page
title: Getting Started
---

## Installing

Grab the latest binary release for your platform from the [Releases]({{ site.github.releases_url }}) page.
You will probably want to copy the binaries into your `PATH`.
This will look something like this:

```bash
VERSION=v0.2.0
curl -sSL {{ site.github.releases_url }}/download/${VERSION}/mcm-${VERSION}-linux_amd64.zip > mcm.zip
unzip mcm.zip -d mcm
cp mcm/* /usr/local/bin/
```

### Building from source

If mcm binaries are not available for your platform, then you may still be able to build from source.
Navigate over to [Releases]({{ site.github.releases_url }}) to download the latest source archive.
Ensure that you have a modern C++ compiler installed (typically through your operating system's package manager).
Then, kick off a [Bazel](https://bazel.build/) build (don't worry, it will be downloaded for you):

```bash
cd path/to/mcm
./bazel build //...

# Optional: run all tests
./bazel test //...
```

Now the binaries will be available under `bazel-bin/`.
You can optionally install them into your PATH:

```bash
# (Optional) Build with optimization enabled
./bazel build -c opt //...

# Copy into your PATH
cp bazel-bin/shellify/mcm-shellify bazel-bin/luacat/mcm-luacat bazel-bin/exec/mcm-exec bazel-bin/dot/mcm-dot /usr/local/bin/
```

## Writing a Catalog

The fundamental concept in mcm is the **catalog**.
The catalog is a binary file that lists the desired end state of your system: what files should exist and what programs to run.
Its structure is defined in [`catalog.capnp`]({{ site.github.repository_url }}/blob/master/catalog.capnp), a [Cap'n Proto schema](https://capnproto.org/language.html).
The catalog is data-driven and purposefully does not define any control flow or variable expansion.
Instead, it is expected that tools that produce a catalog will provide these features.
The main tool that mcm provides to generate a catalog is [mcm-luacat]({{ site.github.repository_url }}/tree/master/luacat/).

mcm-luacat is a small [Lua](https://www.lua.org/) interpreter that acts as a domain-specific language for describing catalogs.
A simple example:

```lua
mcm.resource(
  "hello",  -- name of resource
  {},       -- list of dependencies (none)
  -- The table argument to mcm.file specifies the fields in catalog.capnp.
  mcm.file{
    path = "/etc/hello.txt",
    plain = {content = "Hello, World!\n"},
  }
)

mcm.resource(
  "apt-get update",
  {"hello"},  -- this resource depends on "hello" finishing successfully
  mcm.exec{
    command = {
      argv = {"/usr/bin/apt-get", "update"},
    },
  }
)
```

As you can see, the script follows the schema very closely.
More advanced catalog scripts can use string concatenation, functions, reusable libraries, and other language features of Lua.
To convert this into a catalog file, you use `mcm-luacat`:

```bash
mcm-luacat foo.lua > foo.out
```

Now what can we do with this file?

## Running locally

This is the simplest use case: we want to evaluate and run the catalog on the local machine.

```bash
mcm-exec foo.out
# or sudo mcm-exec foo.out to run the apt-get command
```

Without arguments, each of these tools operates on stdin and stdout, so we can use shell pipelines to simplify this.

```bash
mcm-luacat foo.lua | sudo mcm-exec
```

You may be asking, why is this two separate programs?
Notice that this separation of commands allows you to run the evaluation phase without elevated privileges.
But there's another benefit...

## Running remotely

You can instead generate a script of commands to run remotely.

```bash
mcm-luacat foo.lua | mcm-shellify | ssh root@myrouter.local -c 'bash'
```

While this may not be the exact workflow you would want to use in production (allowing ssh access to root is generally a bad idea), it illustrates how separating out evaluation of a catalog from its application can enable more powerful workflows.
You can imagine storing the bash script on network storage, then executing it on many machines that don't need to have mcm installed.
You could use the bash script in a Docker image.
The point is, any part of the above pipeline can be swapped out for a different program that best suits your needs.

- - -

These are the basics of mcm.
Try it for yourself and get involved with [the project]({{ site.github.repository_url }}).
