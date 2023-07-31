package main

import (
	"reflect"

	"github.com/go-redis/redis"
)

// redisScanAll helps fetching all keys with pagination
func redisScanAll(
	rdb *redis.Client,
	pattern string,
	cursor uint64,
	limit int,
	fetched map[string]bool,
) ([]string, error) {
	vals := []string{}

	if rdb == nil {
		return vals, nil
	}

	keys, cursor, err := rdb.Scan(cursor, pattern, int64(limit)).Result()
	if err != nil {
		return vals, err
	}

	toFetch := []string{}
	for _, k := range keys {
		// no need to refetch the same key
		if _, ok := fetched[k]; ok {
			continue
		}

		fetched[k] = true
		toFetch = append(toFetch, k)
	}

	valsRaw, err := rdb.MGet(toFetch...).Result()
	if err != nil {
		return vals, err
	}

	for _, v := range valsRaw {
		if reflect.TypeOf(v).String() == "string" {
			vals = append(vals, v.(string))
		}
	}

	// this must mean there are more saved
	if len(keys) >= limit {
		res, err := redisScanAll(rdb, pattern, cursor, limit, fetched)
		if err != nil {
			return vals, err
		}

		vals = append(vals, res...)
	}

	return vals, nil
}
