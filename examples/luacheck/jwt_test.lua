require("jwt")

jstr = '{"alg": "ES256", "notsure": "ok"}';
b64jstr = 'eyJhbGciOiAiRVMyNTYiLCAibm90c3VyZSI6ICJvayJ9';
key = "alg"
val = "ES256"

function bad(somestr)
  return "BAD" .. somestr .. "BAD"
end

function test_val(s, key, val, res)
	if has_value(s, key, val) ~= res then
		print("failed: want", res, "for", s, key, val)
	else
		print("success")
	end
end

function tval(fn, s, key, val, res)
	if fn(s, key, val) ~= res then
		print(fn, "failed: want", res, "for", s, key, val)
	else
		print(fn, "success")
	end
end

tval(has_value, b64jstr, key, val, true)
tval(has_value, bad(b64jstr), key, val, false)
-- tval(has_value, b64jstr, bad(key), val, false)
-- tval(has_value, b64jstr, key, bad(val), false)
tval(has_value, " ab c 1 #@!", key, bad(val), false)

-- {"alg": "\u0045S256", "notsure": "ok"}
b64unistr="eyJhbGciOiAiXHUwMDQ1UzI1NiIsICJub3RzdXJlIjogIm9rIn0K"
tval(has_value, b64unistr, key, val, true)

-- {"alg": "\u0045S25\u0036", "notsure": "ok"}
b64unistr="eyJhbGciOiAiXHUwMDQ1UzI1XHUwMDM2IiwgIm5vdHN1cmUiOiAib2sifQo"
tval(has_value, b64unistr, key, val, true)

-- {"alg": "\u0045S25\u0055", "notsure": "ok"}
b64unistr="eyJhbGciOiAiXHUwMDQ1UzI1XHUwMDU1IiwgIm5vdHN1cmUiOiAib2sifQo"
tval(has_value, b64unistr, key, val, false)

tval(has_jwt,b64jstr .. ".xyz", key, val, true)
tval(has_jwt,b64jstr, key, val, false)


-- stub out request_handle and headers
Headers = {}
function Headers:new(o)
  self.headers = o -- or { self.headers = {} }
  setmetatable(o, self)
  self.__index = self
  return self
end

function Headers:get(header_name)
  print(self.headers[header_name], header_name)
  return self.headers[header_name]
end

Request_Handle = {}
function Request_Handle:new (o)
  -- o = o or {}   -- create object if user does not provide one
  self.hdrs = o
  self.resphdr = nil
  self.msg = nil
  setmetatable(o, self)
  self.__index = self
  return self
end
function Request_Handle:respond(resphdr, msg)
  self.resphdr = resphdr
  self.msg = msg
end
function Request_Handle:headers()
  return self.hdrs
end


function testFilter(hname, jstr, reject)
  hh = {}
  hh[hname] = jstr .. ".BBB.SSS"
  -- hdr = Headers:new({Authorization="Bearer " .. jstr .. ".BBB.SSS"})
  hdr = Headers:new(hh)

  rej = should_reject_request_by_headers(hdr)
  if rej ~= reject then
    print("testFilter failed reject. want =", reject, "got", rej)
  else
    print("testFilter success")
  end
end

-- case 1
-- {"alg": "ES256", "notsure": "ok"}
testFilter("Authorization", "Bearer " .. b64jstr, true)

-- {"alg": "RSA", "notsure": "ok"}
b64jstr2 = "eyJhbGciOiAiUlNBIiwgIm5vdHN1cmUiOiAib2sifQo="
testFilter("Authorization", "Bearer " .. b64jstr2, false)

-- send ES256 thru an alternate headers, but do not configure x-my-header as headersToCheck
testFilter("x-my-header", b64jstr, false)

headersToCheck = {"x-my-header"}
-- same as above after configuring headersToCheck, should reject request now
testFilter("x-my-header", b64jstr, true)
testFilter("x-my-header", b64jstr2, false)


--- url parse tests
function test_find_qp(path, qp, qv)
  qva = find_qp(path, qp)
  if qva ~= qv then
    print(path, qp, "want", qv, "got", qva)
  else
    print("success")
  end
end

test_find_qp("/uri?k1=v1&k2=v2", "k1", "v1")
test_find_qp("/uri?k1=v1&k2=vv2", "k2", "vv2")
test_find_qp("/uri?k1=v1&k2=v2", "k3", nil)
test_find_qp("/uri", "k3", nil)
test_find_qp("/abc?x-my-header=eyJhbGciOiAiUlNBIiwgIm5vdHN1cmUiOiAib2sifQo=.BBB.SSS", "x-my-header", "eyJhbGciOiAiUlNBIiwgIm5vdHN1cmUiOiAib2sifQo=.BBB.SSS")

function testFilter_qp(qp, jstr, reject)
  rej = should_reject_request_by_query_params("/abc?".. qp .. "=" .. jstr .. ".BBB.SSS")
  if rej ~= reject then
    print("testFilter_qp failed reject. want =", reject, "got", rej)
  else
    print("testFilter_qp success")
  end
end

-- {"alg": "RSA", "notsure": "ok"}
b64jstr2 = "eyJhbGciOiAiUlNBIiwgIm5vdHN1cmUiOiAib2sifQo="

-- send ES256 thru an alternate headers, but do not configure x-my-header as headersToCheck
testFilter_qp("myqueryparam", b64jstr, false)

paramsToCheck = {"myqueryparam"}
-- same as above after configuring headersToCheck, should reject request now
testFilter_qp("myqueryparam", b64jstr, true)
testFilter_qp("myqueryparam", b64jstr2, false)
paramsToCheck = {"my-query-param"}
testFilter_qp("my-query-param", b64jstr, true)

