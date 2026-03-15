package limiter

const luaScriptSource = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local window_start = now - window

local fields = redis.call("HGETALL", key)
local total = 0

for i = 1, #fields, 2 do
    local bucket_ts = tonumber(fields[i])
    if bucket_ts <= window_start then
        redis.call("HDEL", key, fields[i])
    else
        total = total + tonumber(fields[i+1])
    end
end

if total >= limit then
    return 0
else
    redis.call("HINCRBY", key, ARGV[1], 1)
    redis.call("EXPIRE", key, window + 2)
    return 1
end
`
