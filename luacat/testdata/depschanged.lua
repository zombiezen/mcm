-- "xyzzy!" is specially picked to have the highest bit set high
mcm.resource("xyzzy!", {}, mcm.file{
  path = "/etc/motd",
  plain = {},
})

mcm.resource("apt-get update", {"xyzzy!"}, mcm.exec{
  condition = {ifDepsChanged = {0xd96f419065c49db1}},
  command = {
    argv = {"/usr/bin/apt-get", "update"},
  },
})
