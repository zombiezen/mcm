-- Copyright 2017 The Minimal Configuration Manager Authors
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

require("string")
local table = require("table")

local shlib = {}

function shlib.quote(s)
  if s == "" then
    return "''"
  end
  local unsafePat = "[^-_/.A-Za-z0-9]"
  if not s:find(unsafePat) then
    return s
  end
  local parts = {"'"}
  local i = 1
  while i <= #s do
    local j = s:find("'", i)
    if not j then
      parts[#parts+1] = s:sub(i)
      break
    end
    parts[#parts+1] = s:sub(i, j-1)
    parts[#parts+1] = "'\\''"
    i = j + 1
  end
  parts[#parts+1] = "'"
  return table.concat(parts)
end

return shlib
