-- This script checks jwt token header for a specific alg
-- If this is found then it rejects the request.
-- has_value(b64jstr, key, val) checks if b64jstr has key==val
-- auth_header -- get jwt token from Authorization: Bearer
-- check_headers -- get jwt token from custom headers
-- check_params -- get jwt token from custom params

-- implementation notes
-- base64encode / decode is from libresolv

ffi = require("ffi")
ffi.cdef[[
  void free(void *ptr);
  void *malloc(size_t size);
]]

resolv = ffi.load("/lib/x86_64-linux-gnu/libresolv.so.2", false)
ffi.cdef[[
typedef    unsigned char    u_char;

int __b64_ntop(u_char const *src, size_t srclength, char *target, size_t targsize);
int __b64_pton (char const *src, u_char *target, size_t targsize);
]]

-- has_value(b64jstr, key, val) checks if b64jstr has key==val
function has_value(b64jstr, key, val)
  -- at most decoded string will be equal to b64 string.

  local decoded = ffi.new("unsigned char[".. b64jstr:len() .. "]")

  local szd = resolv.__b64_pton(b64jstr, decoded, b64jstr:len())
  if szd == -1 then return false end
  local decodedS = ffi.string(decoded, szd)

  -- look for "alg":"ES256"
  -- return decodedS:match('"alg"%s*:%s*"ES256"') ~= nil
  return decodedS:match('"' .. key .. '"%s*:%s*"' .. val .. '"') ~= nil
end

-- has_jwt if the given token header can be decoded,
-- check if header[key] == val
-- token <b64 header>.<b64 body>.<b64 signature>
function has_jwt(token, key, val)
  if token == nil or token:len() == 0 then return false end

  local idx, idx1 = string.find(token, "%.");
  if idx == nil then return false end

  local token_header = string.sub(token, 0, idx-1)

  return has_value(token_header, key, val)
end

-- jwt token typically uses
-- Authorization: Bearer <hhh.bbb.sss>
function check_auth_header(headers, key, val)
  authh = headers:get("Authorization")

  if authh == nil then return false end

  first = true
  for w in authh:gmatch("%S+") do
    if first then
      first = false
      -- if this is not bearer token, we don't care
      if w ~= "Bearer" and w ~= "bearer" then
        return false
      end
    else
      return has_jwt(w, key, val)
    end
  end
end

findKey = "alg"
findVal = "ES256"

-- if you use custom headers to carry jwt tokens list them here.
-- jwt token will be extracted from them.
headersToCheck = {}

-- if you use custom query parameters to carry jwt tokens list them here.
-- jwt token will be extracted from them.
paramsToCheck = {}

function should_reject_request_by_query_params(path)
  if path == nil then return false end

  for _, qp_name in ipairs(paramsToCheck) do
    -- if qp_name has "-" in this, it is special martching char for pattern in find
    -- so we escape it
    qp_name = string.gsub(qp_name, "%-", "%%-")
    authh = find_qp(path, qp_name)
    if has_jwt(authh, findKey, findVal) then
      return true
    end
  end

  return false
end

function should_reject_request_by_headers(headers)
  -- check most common case first authorization header
  if check_auth_header(headers, findKey, findVal) then
    return true
  end

  for _, header_name in ipairs(headersToCheck) do
    authh = headers:get(header_name)
    if has_jwt(authh, findKey, findVal) then
      return true
    end
  end

  return false
end

-- given path find a query parameter denoted by qp
-- uri?k1=v1&k2=v2&...
function find_qp(path, qp)
	local start, _ = string.find(path, "?")

	if start == nil then return nil end

	local qp_start, qp_end = string.find(path, qp .. "=", start + 1)

	if qp_start == nil then return nil end

	-- found qp param
	local qv_end, _ = string.find(path, "&", qp_end+1)

	-- this is the last value, cannot find a "&"
	if qv_end == nil then return string.sub(path, qp_end+1) end

	return string.sub(path, qp_end+1, qv_end-1)
end

-- envoy entry point
function envoy_on_request(request_handle)
  headers = request_handle:headers()

  if should_reject_request_by_headers(headers) then
    request_handle:respond({[":status"] = "401"}, "bad jwt value")
  end

  if should_reject_request_by_query_params(headers:get(":path")) then
    request_handle:respond({[":status"] = "401"}, "bad jwt value")
  end
end
