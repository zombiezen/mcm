-- A basic-ish catalog.

local samplelib = require 'samplelib'

local function filepath(name)
  return '/tmp/mcmtest/'..name..'-mcm.txt'
end

local homedir = samplelib.mkparent('/tmp/mcmtest', '/tmp')

mcm.resource('foo', {mcm.hash('bar'), homedir}, mcm.file{
  path = filepath('foo'),
  plain = {
    content = 'Hello, World!',
  },
})

mcm.resource('bar', {homedir}, mcm.file{
  path = filepath('bar'),
  plain = {
    content = 'Good bye, World!',
  },
})

mcm.resource('apt-get update', {}, mcm.exec{
  command = {
    argv = {"/usr/bin/apt-get", "update"},
  },
})
