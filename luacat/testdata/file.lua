mcm.resource(
  "hello",  -- name of resource
  {},       -- list of dependencies (none)
  -- The table argument to mcm.file specifies the fields in catalog.capnp.
  mcm.file{
    path = "/etc/hello.txt",
    plain = {content = "Hello, World!\n"},
  }
)
