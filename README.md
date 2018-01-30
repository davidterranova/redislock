# redis lock
Simple distributed lock implementation based on [redis distlock](https://redis.io/topics/distlock)

## Dependance
Redis client connection [github.com/go-redis/redis](https://github.com/go-redis/redis)

## Usage
```golang
var ctx, _ = context.WithTimeout(context.Background(), time.Duration(1) * time.Second)
var locker = NewLocker(client, "mutex_key")

if err := locker.Lock(ctx); err == ErrAlreadyLocked {
  // lock already taken and not able to acquire it in the allowed time
} else {
  // lock acquired
  // do stuff ...
  
  // release
  locker.Unlock()
}
```