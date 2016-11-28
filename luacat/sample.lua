-- Copyright 2016 The Minimal Configuration Manager Authors
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

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
