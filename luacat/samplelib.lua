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

local samplelib = {}

local function hasPrefix(s, prefix)
  if prefix == '' then return true end
  return s:sub(1, #prefix) == prefix
end

local function hasSuffix(s, suffix)
  if suffix == '' then return true end
  return s:sub(-#suffix) == suffix
end

function samplelib.mkparent(path, prefix)
  if not hasPrefix(prefix, '/') then
    error('mkparent: prefix must be an absolute path, got '..prefix)
  end
  if not hasSuffix(prefix, '/') then
    prefix = prefix .. '/'
  end
  if not hasPrefix(path, prefix) then
    error('mkparent: path "'..path..'" does not start with prefix '..prefix)
  end

  -- Find directory parts
  local parts = {}
  local start = #prefix+1
  repeat
    local i = path:find('/', start)
    if i then
      parts[#parts+1] = path:sub(1, i-1)
      start = i+1
    else
      parts[#parts+1] = path
    end
  until parts[#parts] == path

  -- Create resources
  local idPrefix = 'mkparent:'
  local lastID
  for _, dir in ipairs(parts) do
    local curr = mcm.hash(idPrefix..dir)
    mcm.resource(curr, {lastID}, mcm.file{
      path = dir,
      directory = {},
    })
    lastID = curr
  end

  -- Return the deepest directory's ID
  return lastID
end

return samplelib
