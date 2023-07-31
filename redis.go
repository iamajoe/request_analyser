package main

import (
	"reflect"
	"strings"

	"github.com/go-redis/redis"
)

// redisScanKeys helps fetching keys with pagination
func redisScanKeys(
	rdb *redis.Client,
	pattern string,
	cursor uint64,
	limit int,
) ([]string, uint64, error) {
	if rdb == nil {
		return []string{}, cursor, nil
	}

	keys, cursor, err := rdb.Scan(cursor, pattern, int64(limit)).Result()
	if err != nil {
		return keys, cursor, err
	}

	return keys, cursor, nil
}

// redisGetsKeys fetches a set of keys
func redisGetKeys(
	rdb *redis.Client,
	keys []string,
) ([]string, error) {
	vals := []string{}

	if rdb == nil {
		return vals, nil
	}

	valsRaw, err := rdb.MGet(keys...).Result()
	if err != nil {
		return vals, err
	}

	for _, v := range valsRaw {
		if v != nil && reflect.TypeOf(v).String() == "string" {
			vals = append(vals, v.(string))
		}
	}

	return vals, nil
}

// redisScanAll fetches to a file all data
func redisScanAll(
	rdb *redis.Client,
	pattern string,
	writer func(raw string) error,
) error {
	fetched := make(map[string]bool)

	var keys []string
	var err error
	cursor := uint64(0)
	limit := 1000

	// continue to fetch until we have
	for len(keys) >= limit-10 || len(fetched) == 0 {
		keys, cursor, err = redisScanKeys(rdb, pattern, cursor, limit)
		if err != nil {
			return err
		}

		parsedKeys := []string{}

		// go per key, we want to make sure we get them
		for _, k := range keys {
			// already fetched, do not add it again
			if _, ok := fetched[k]; ok {
				continue
			}

			// cache the key
			fetched[k] = true
			parsedKeys = append(parsedKeys, k)

		}

		// we now have keys, we want to fetch and append to file
		vals, err := redisGetKeys(rdb, parsedKeys)
		if err != nil {
			return err
		}

		raw := strings.Join(vals, "\n")
		if err := writer(raw); err != nil {
			return err
		}
	}

	return nil
}
