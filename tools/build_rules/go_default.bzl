load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

_go_default_rule_name = "go_default_library"


def go_default_library(
    name=_go_default_rule_name,
    data=[],
    deps=[],
    exclude=[],
    test=False,
    test_data=[],
    test_deps=[],
    test_separate=False,
    test_size="small",
    testonly=False,
    visibility=None):
  srcspat = "*.go"
  testpat = "*_test.go"
  testdatapat = "testdata/**/*"
  if name != _go_default_rule_name:
    srcspat = name + "/" + srcspat
    testpat = name + "/" + testpat
    testdatapat = name + "/" + testdatapat
  go_library(
      name = name,
      srcs = native.glob([srcspat], [testpat] + exclude),
      data = data,
      deps = deps,
      testonly = testonly,
      visibility = visibility,
  )
  if test and test_separate:
    go_test(
        name = name + "_test",
        srcs = native.glob([testpat], exclude),
        data = test_data + native.glob([testdatapat], exclude),
        size = test_size,
        deps = test_deps,
        visibility = visibility,
    )
  elif test and not test_separate:
    go_test(
        name = name + "_test",
        srcs = native.glob([testpat], exclude),
        data = test_data + native.glob([testdatapat], exclude),
        library = ":" + name,
        size = test_size,
        deps = test_deps,
        visibility = visibility,
    )
