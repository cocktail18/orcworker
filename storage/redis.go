package storage

import (
	"github.com/cocktail18/orcworker"
	"github.com/go-redis/redis"
	"strconv"
	"github.com/vmihailenco/msgpack"
)

const (
	QUEUE_KEY = "__orcworker_queue"
	SET_KEY   = "__orcworker_set"
)

type RedisStorage struct {
	conn *redis.Client
}

func NewRedisStorage(host string, port int, password string, db int) (*RedisStorage, error) {
	m := &RedisStorage{}
	opts := &redis.Options{
		Addr:     host + ":" + strconv.Itoa(port),
		Password: password,
		DB:       db,
	}
	m.conn = redis.NewClient(opts)
	err := m.conn.Ping().Err()
	return m, err
}

func (redisStorage *RedisStorage) EnQueue(seed *orcworker.Seed) error {
	b, err := msgpack.Marshal(seed)
	if err != nil {
		return err
	}
	hash, err :=  seed.Sha256()
	if err != nil {
		return err
	}
	pipe := redisStorage.conn.Pipeline()
	pipe.SAdd(SET_KEY, hash)
	pipe.LPush(QUEUE_KEY, string(b))
	_, err = pipe.Exec()
	return err
}

func (redisStorage *RedisStorage) DeQueue() (*orcworker.Seed, error) {
	str, err := redisStorage.conn.RPop(QUEUE_KEY).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, orcworker.ERR_SEEDS_EMPTY
		}
		return nil, err
	}
	var seed orcworker.Seed
	err = msgpack.Unmarshal([]byte(str), &seed)
	return &seed, err
}

func (redisStorage *RedisStorage) QueueCapacity() (int, error) {
	l, err := redisStorage.conn.LLen(QUEUE_KEY).Result()
	return int(l), err
}

func (redisStorage *RedisStorage) IsContain(seed *orcworker.Seed) (bool, error) {
	hash, err := seed.Sha256()
	if err != nil {
		return false, err
	}
	return redisStorage.conn.SIsMember(SET_KEY, hash).Result()
}
