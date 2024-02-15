"""Redis.

Redis is an open-source, in-memory data structure store, used as a database, cache, and message broker.

References:
  - Redis command reference: https://redis.io/commands/
  - Go client API: https://pkg.go.dev/github.com/go-redis/redis/v9
"""

def blpop(timeout, *keys):
    pass

def brpop(timeout, *keys):
    pass

def brpoplpush(src, dst):
    pass

def decr(key):
    pass

def decrby(key, by):
    pass

def delete(*keys):
    pass

def do():
    pass

def expire(key, expiration, nx, xx, gt, lt):
    pass

def get(key):
    pass

def incr(key):
    pass

def incrby(key, by):
    pass

def lindex(key, index):
    pass

def linsert(key, op, pivot, value):
    pass

def llen(key):
    pass

def lpop(key):
    pass

def lpos(key, value, rank, max_len):
    pass

def lpush(key, *vs):
    pass

def lpushx(key, *vs):
    pass

def lrange(key, start, stop):
    pass

def lrem(key):
    pass

def lset(key, index, value):
    pass

def ltrim(key, start, stop):
    pass

def rpop(key):
    pass

def rpoplpush(src, dst):
    pass

def rpush(key, *vs):
    pass

def rpushx(key, *vs):
    pass

def set(key, value, ttl):
    pass