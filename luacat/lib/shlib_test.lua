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

quoteTests = {
  {input="", output="''"},
  {input="abc", output="abc"},
  {input="abc def", output="'abc def'"},
  {input="abc/def", output="abc/def"},
  {input="abc.def", output="abc.def"},
  {input="\"abc\"", output="'\"abc\"'"},
  {input="'abc'", output="''\\''abc'\\'''"},
  {input="abc\\", output="'abc\\'"},
}

local shlib = require("shlib")
local os = require("os")
local string = require("string")

local ok = true
for _, t in ipairs(quoteTests) do
  local out = shlib.quote(t.input)
  if out ~= t.output then
    print(string.format("shlib.quote(%q) = %q; want %q", t.input, out, t.output))
    ok = false
  end
end
if not ok then
  os.exit(false)
end
