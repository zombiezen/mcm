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

local apt = {
  curlID = mcm.hash('apt.curl'),
  gdebiID = mcm.hash('apt.gdebi'),
  updateID = mcm.hash('apt.update'),
}
local aptConfigs = {}
local aptGetPath = "/usr/bin/apt-get"
local dpkgPath = "/usr/bin/dpkg"
local dpkgQueryPath = "/usr/bin/dpkg-query"
local header = "# This file is managed by mcm. DO NOT EDIT.\n"
local execEnv = {
  {name = "DEBIAN_FRONTEND", value = "noninteractive"},
  {name = "PATH", value = "/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin"},
}

local function addTable(t1, ...)
  local t = table.move(t1, 1, #t1, 1, {})
  local n = select("#", ...)
  for i = 1, n do
    local ti = select(i, ...)
    table.move(ti, 1, #ti, #t+1, t)
  end
  return t
end

local function aptpkg(id, deps, args)
  if args.package == nil then error("must pass [\"package\"] to apt.pkg") end

  local queryCmd = {
    bash = string.format("%s --status '%s' > /dev/null", dpkgQueryPath, args.package),
    environment = execEnv,
  }
  local e
  if args.install == nil or args.install then
    e = mcm.exec{
      command = {
        argv = {aptGetPath, "install", "-y", "--no-install-recommends", args.package},
        environment = execEnv,
      },
      condition = {unless = queryCmd},
    }
  else
    e = mcm.exec{
      command = {
        argv = {aptGetPath, "remove", "-y", args.package},
        environment = execEnv,
      },
      condition = {onlyIf = queryCmd},
    }
  end
  mcm.resource(id, deps, e)
end

local function aptpin(id, deps, args)
  if type(args.enable) ~= "boolean" then error("must pass [\"enable\"] boolean to apt.pin") end
  if type(args.name) ~= "string" then error("must pass [\"name\"] string to apt.pin") end
  local path = "/etc/apt/preferences.d/"..args.name..".pref"

  if not args.enable then
    mcm.resource(id, deps, mcm.file{
      path = path,
      absent = true,
    })
  end

  if type(args.explanation) ~= "string" then error("must pass [\"explanation\"] string to apt.pin") end
  local packages = args.packages or "*"
  if type(args.release) ~= "string" then error("must pass [\"release\"] string to apt.pin") end
  local priority = args.priority or 0
  local pin = "release a="..args.release

  local content = header..string.format([[Explanation: %s
Package: %s
Pin: %s
Pin-Priority: %d
]], args.explanation, packages, pin, args.priority)
  -- TODO(soon): ensure root/root 0644
  mcm.resource(id, deps, mcm.file{
    path = path,
    plain = {content = content},
  })
end

local function aptsource(id, deps, args)
  if type(args.enable) ~= "boolean" then error("must pass [\"enable\"] boolean to apt.source") end
  if type(args.name) ~= "string" then error("must pass [\"name\"] string to apt.source") end
  local path = "/etc/apt/sources.list.d/"..args.name..".list"

  if not args.enable then
    mcm.resource(id, deps, mcm.file{
      path = path,
      absent = true,
    })
    return
  end

  local repos = args.repos or "main"
  if type(args.location) ~= "string" then error("must pass [\"location\"] string to apt.source") end
  if type(args.release) ~= "string" then error("must pass [\"release\"] string to apt.source") end
  local include = args.include or {deb=true, src=false}
  local line = args.location.." "..args.release.." "..repos.."\n"
  if args.architecture then
    line = "[arch="..args.architecture.."] "..line
  end
  local content = header
  if include.deb then
    content = content.."deb "..line
  end
  if include.src then
    content = content.."deb-src "..line
  end
  mcm.resource(id, deps, mcm.file{
    path = path,
    plain = {content = content},
  })
end

function apt.pkg(id, deps, args)
  aptpkg(id, addTable(deps, {apt.updateID}), args)
end

function apt.pin(id, deps, args)
  if type(id) == "string" then
    id = mcm.hash(id)
  end
  aptpin(id, deps, args)
  aptConfigs[#aptConfigs + 1] = id
end

function apt.source(id, deps, args)
  if type(id) == "string" then
    id = mcm.hash(id)
  end
  aptsource(id, deps, args)
  aptConfigs[#aptConfigs + 1] = id
end

function apt.urldeb(id, deps, package, version, url)
  local packagePath = "/var/cache/apt/archives/"..package.."_"..version..".deb"
  mcm.resource(id, addTable(deps, {apt.gdebiID, apt.curlID}), mcm.exec{
    command = {
      bash = string.format("/usr/bin/curl -sSL '%s' > '%s' && /usr/bin/gdebi --non-interactive '%s'", url, packagePath, packagePath),
      environment = execEnv,
    },
    condition = {
      unless = {
        bash = string.format("%s --status '%s' > /dev/null && %s --compare-versions $(%s --show -f '${Version}' '%s') ge '%s'", dpkgQueryPath, package, dpkgPath, dpkgQueryPath, package, version),
        environment = execEnv,
      },
    },
  })
end

function apt.finish()
  if #aptConfigs == 0 then
    mcm.resource(apt.updateID, {}, mcm.noop)
    return
  end
  mcm.resource(apt.updateID, aptConfigs, mcm.exec{
    condition = {ifDepsChanged = aptConfigs},
    command = {
      argv = {aptGetPath, "update"},
      environment = execEnv,
    },
  })
end

-- Initialization
do
  aptpkg(apt.curlID, {}, {package = "curl"})
  local gdebiPkg = mcm.hash("gdebi package")
  local aptTransportPkg = mcm.hash("apt-transport-https package")
  aptpkg(gdebiPkg, {}, {package = "gdebi"})
  aptpkg(aptTransportPkg, {}, {package = "apt-transport-https"})
  mcm.resource(apt.gdebiID, {gdebiPkg, aptTransportPkg}, mcm.noop)
end
return apt
