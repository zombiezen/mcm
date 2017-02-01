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

local string = require("string")
local mcm = require("mcm")
local shlib = require("shlib")
local table = require("table")

local configs = {}

local grepPath = "/bin/grep"
local sedPath = "/bin/sed"

function configs.escapeRegex(s)
  local parts = {}
  local i = 1
  while i <= #s do
    -- POSIX.2 BRE special characters are: .[\*^$
    -- We also escape slash so that it can be used as a sed address.
    local j = s:find("[/.[\\*^$]", i)
    if not j then
      parts[#parts+1] = s:sub(i)
      break
    end
    parts[#parts+1] = s:sub(i, j-1)
    parts[#parts+1] = "\\"
    parts[#parts+1] = s:sub(j, j)
    i = j + 1
  end
  return table.concat(parts)
end

function configs.line(id, deps, args)
  if type(args.path) ~= "string" then error("must pass [\"path\"] string to configs.line") end
  if type(args.pattern) ~= "string" then error("must pass [\"pattern\"] string to configs.line") end
  if args.add and args.remove then error("must pass only one of [\"add\"] or [\"remove\"] to configs.line") end
  if not args.add and not args.remove then error("must pass one of [\"add\"] or [\"remove\"] to configs.line") end
  -- TODO(someday): add header for "managed by mcm"

  if args.add then
    -- TODO(darwin): -i requires a suffix and needs cleanup
    local script = string.format([=[
p=%s
if [ -e "$p" ]; then
  %s -i -e /%s/d "$p" || exit 1
fi
echo %s >> "$p"
]=], shlib.quote(args.path), shlib.quote(sedPath), shlib.quote(args.pattern), shlib.quote(args.add))
    mcm.resource(id, deps, mcm.exec{
      command = {
        bash = script,
      },
      condition = {
        unless = {
          argv = {grepPath, "^"..configs.escapeRegex(args.add).."$", args.path},
        },
      },
    })
  else
    mcm.resource(id, deps, mcm.exec{
      command = {
        -- TODO(darwin): -i requires a suffix and needs cleanup
        argv = {sedPath, "-i", "-e", "/"..args.pattern.."/d", args.path},
      },
      condition = {
        onlyIf = {
          argv = {grepPath, args.pattern, args.path},
        },
      },
    })
  end
end

function configs.host(id, deps, args)
  if type(args.name) ~= "string" then error("must pass [\"name\"] string to configs.host") end

  local hostFile = "/etc/hosts"
  -- Match on the second field (name)
  local pat = "^[ \t]*[^ \t#]\\{1,\\}[ \t]\\{1,\\}"..configs.escapeRegex(args.name).."\\([ \t].*\\)\\{0,1\\}$"
  if not args.absent then
    if type(args.ip) ~= "string" then error("must pass [\"ip\"] string to configs.host") end
    local fields = {args.ip, args.name}
    if args.aliases then
      for _, a in ipairs(args.aliases) do
        fields[#fields+1] = a
      end
    end
    configs.line(id, deps, {
      path = hostFile,
      add = table.concat(fields, "\t"),
      pattern = pat,
    })
  else
    configs.line(id, deps, {
      path = hostFile,
      remove = true,
      pattern = pat,
    })
  end
end

return configs
