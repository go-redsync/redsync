package redsync

import "github.com/go-redsync/redsync/v4/redis"

var deleteScript = redis.NewScript(1, `
	local val = redis.call("GET", KEYS[1])
	if val == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	elseif val == false then
		return -1
	else
		return 0
	end
`, "e950836ed1e694540c503ef9972b8de518044d3b")

var touchWithSetNXScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	elseif redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2], "NX") then
		return 1
	else
		return 0
	end
`, "b6b55c439e02c6f52ffb8be7ec8c0d0ac816ea79")

var touchScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	else
		return 0
	end
`, "d75bb8b5bd13532ddedd17762e772346e39b321e")
