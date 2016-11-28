# Copyright 2016 The Minimal Configuration Manager Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Cap'n Proto"""

load("@io_bazel_rules_go//go:def.bzl", "go_library")


capnp_filetype = FileType([".capnp"])


def _symlink_tree_commands(dest_dir, artifact_dict):
  """Symlink_tree_commands returns a list of commands to create the
  dest_dir, and populate it according to the given dict.

  Args:
    dest_dir: The destination directory, a string.
    artifact_dict: The mapping of (path in the dest_dir) => (File).

  Returns:
    A list of commands that will setup the symlink tree.
  """
  cmds = [
    "rm -rf " + dest_dir,
    "mkdir -p " + dest_dir,
  ]

  for item in artifact_dict.items():
    new_path = item[0]
    old_path = item[1].path
    up = dest_dir.count('/') + 1
    i = new_path.rfind('/')
    if i != -1:
      new_dir = new_path[:i]
      up += new_path.count('/')
      cmds += ["mkdir -p '%s/%s'" % (dest_dir, new_dir)]
    cmds += ["ln -s %s%s '%s/%s'" % ('../' * up, old_path, dest_dir, new_path)]

  return cmds


def _collect_schemas(ctx, src, deps):
  schemas = set()
  for dep in deps:
    schemas += dep.capnp_schemas
  schemas += [src]
  return schemas


def _build_filemap(schemas):
  filemap = {}
  for s in schemas:
    filemap[s.short_path] = s
  return filemap


def _capnp_schema_impl(ctx):
  src = ctx.file.src
  deps = ctx.attr.deps
  out = ctx.outputs.capnp_out
  capnp_tool = ctx.executable._capnp_tool
  transitive_schemas = _collect_schemas(ctx, src, deps)

  includes_dir = out.path + '.includes'
  cmds = _symlink_tree_commands(
      includes_dir, _build_filemap(transitive_schemas))
  args = [capnp_tool.path, "compile", "-o-",
          "-I" + includes_dir,
          "--no-standard-import",
          "--src-prefix=" + src.root.path,
          src.path]
  cmds += [" ".join(args) + " > " + out.path]
  ctx.action(
      inputs = list(transitive_schemas) + [capnp_tool],
      outputs = [out],
      mnemonic = "CapnpCompile",
      command = " && ".join(cmds))
  return struct(
      capnp_schemas = transitive_schemas,
      capnp_basename = src.short_path)


_capnp_attrs = {
    "deps": attr.label_list(providers=["capnp_schemas"]),
    "licenses": attr.license(),
    "_capnp_tool": attr.label(
        default = Label("//third_party/capnproto:capnp"),
        cfg = "host",
        executable = True),
}


capnp_library = rule(_capnp_schema_impl,
    attrs = _capnp_attrs + {
        "src": attr.label(allow_files=capnp_filetype, single_file=True),
    },
    outputs = {
        "capnp_out": "%{name}.out",
    },
    output_to_genfiles = True)


def _capnp_gengo_impl(ctx):
  schema = ctx.file.src
  basename = ctx.attr.src.capnp_basename
  out = ctx.outputs.capnp_gosrc
  capnpc_go = ctx.executable._capnpc_tool

  args = ["../" * (ctx.genfiles_dir.path.count("/") + 1) + capnpc_go.path]
  tmpout = ctx.genfiles_dir.path + "/" + basename + ".go"
  ctx.action(
      inputs = [schema, capnpc_go],
      outputs = [out],
      mnemonic = "CapnpGenGo",
      command = "(cd %s && %s) < %s && mv %s %s" % (
          ctx.genfiles_dir.path, " ".join(args), schema.path, tmpout, out.path))
  return struct()


_capnp_gengo = rule(_capnp_gengo_impl,
    attrs = {
        "src": attr.label(providers=["capnp_basename"], allow_single_file=True, mandatory=True),
        "_capnpc_tool": attr.label(
            default = Label("//third_party/golang/capnproto/capnpc-go"),
            cfg = "host",
            executable = True),
    },
    outputs = {
        "capnp_gosrc": "%{name}.go",
    },
    output_to_genfiles = True)


def capnp_go_library(
    name,
    lib,
    deps = [],
    testonly = False,
    visibility = None):
  _capnp_gengo(
      name = name + "_gosrc",
      src = lib,
      testonly = testonly,
  )
  go_library(
      name = name,
      srcs = [":" + name + "_gosrc"],
      testonly = testonly,
      deps = deps + [
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:schemas",
        "//third_party/golang/capnproto:server",
        "//third_party/golang/capnproto/encoding/text:go_default_library",
      ],
      visibility = visibility,
  )


def _capnp_gencc_impl(ctx):
  schema = ctx.file.src
  basename = ctx.attr.src.capnp_basename
  srcout = ctx.outputs.out
  hdrout = ctx.outputs.hdr
  capnpc_cc = ctx.executable._capnpc_tool

  args = ["../" * (ctx.genfiles_dir.path.count("/") + 1) + capnpc_cc.path]
  tmpsrc = ctx.genfiles_dir.path + "/" + basename + ".c++"
  tmphdr = ctx.genfiles_dir.path + "/" + basename + ".h"
  cmd = "(cd %s && %s) < %s" % (ctx.genfiles_dir.path, " ".join(args), schema.path)
  if tmpsrc != srcout.path:
    cmd += " && mv %s %s" % (tmpsrc, srcout.path)
  if tmphdr != hdrout.path:
    cmd += " && mv %s %s" % (tmphdr, hdrout.path)
  ctx.action(
      inputs = [schema, capnpc_cc],
      outputs = [srcout, hdrout],
      mnemonic = "CapnpGenCc",
      command = cmd)
  return struct()


_capnp_gencc = rule(_capnp_gencc_impl,
    attrs = {
        "src": attr.label(providers=["capnp_basename"], allow_single_file=True, mandatory=True),
        "hdr": attr.output(mandatory=True),
        "out": attr.output(mandatory=True),
        "_capnpc_tool": attr.label(
            default = Label("//third_party/capnproto:capnpc-c++"),
            cfg = "host",
            executable = True),
    },
    output_to_genfiles = True)


def capnp_cc_library(
    name,
    lib,
    basename,
    deps = [],
    testonly = False,
    visibility = None):
  if not basename.endswith(".capnp"):
    fail("\"%s\" does not end with .capnp" % (basename), "basename")
  _capnp_gencc(
      name = name + "_ccsrc",
      src = lib,
      hdr = basename + ".h",
      out = basename + ".c++",
      testonly = testonly,
  )
  native.cc_library(
      name = name,
      srcs = [basename + ".c++"],
      hdrs = [basename + ".h"],
      testonly = testonly,
      deps = deps + [
        "//third_party/capnproto:capnp_lib",
      ],
      visibility = visibility,
  )


def _capnp_eval_impl(ctx):
  src = ctx.file.src
  symbol = ctx.attr.symbol
  deps = ctx.attr.deps
  out = ctx.outputs.out
  capnp_tool = ctx.executable._capnp_tool
  transitive_schemas = _collect_schemas(ctx, [src], deps)

  includes_dir = out.path + '.includes'
  cmds = _symlink_tree_commands(
      includes_dir, _build_filemap(transitive_schemas))
  args = [
      ctx.executable._capnp_tool.path,
      "eval",
      "--no-standard-import",
      "-I" + includes_dir,
  ]
  if ctx.attr.binary:
    args += ["--binary"]
  args += [ctx.file.src.path, symbol]
  cmds += [" ".join(args) + " > " + out.path]
  ctx.action(
      inputs = list(transitive_schemas) + [capnp_tool],
      outputs = [out],
      mnemonic = "CapnpEval",
      command = " && ".join(cmds))


capnp_eval = rule(_capnp_eval_impl,
    attrs = _capnp_attrs + {
        "src": attr.label(allow_files=capnp_filetype, single_file=True),
        "out": attr.output(mandatory = True),
        "symbol": attr.string(mandatory = True),
        "binary": attr.bool(default = True),
    },
    output_to_genfiles = True)

# vim: ft=python et ts=8 sts=2 sw=2 tw=0
