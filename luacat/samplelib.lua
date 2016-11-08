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
