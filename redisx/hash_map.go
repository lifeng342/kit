package redisx

import (
	"context"
	"errors"
	"time"

	"github.com/mbeoliero/kit/utils/typex"
	"github.com/redis/go-redis/v9"
)

type HashMap[K comparable, V any] struct {
	Key string
	Cli redis.UniversalClient
}

func NewHashMap[K comparable, V any](cli redis.UniversalClient, key string) *HashMap[K, V] {
	return &HashMap[K, V]{
		Key: key,
		Cli: cli,
	}
}

// Set sets a field in the hash
func (h *HashMap[K, V]) Set(ctx context.Context, field K, value V, expire time.Duration) error {
	pipe := h.Cli.Pipeline()
	pipe.HSet(ctx, h.Key, typex.ToString(field), typex.ToString(value))
	if expire > 0 {
		pipe.Expire(ctx, h.Key, expire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// SetMulti sets multiple fields in the hash
func (h *HashMap[K, V]) SetMulti(ctx context.Context, fields map[K]V, expire time.Duration) error {
	if len(fields) == 0 {
		return nil
	}

	values := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		values[typex.ToString(k)] = typex.ToString(v)
	}

	pipe := h.Cli.Pipeline()
	pipe.HSet(ctx, h.Key, values)
	if expire > 0 {
		pipe.Expire(ctx, h.Key, expire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Get gets a field from the hash
func (h *HashMap[K, V]) Get(ctx context.Context, field K) (V, error) {
	var res V
	val, err := h.Cli.HGet(ctx, h.Key, typex.ToString(field)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return res, nil
		}
		return res, err
	}
	return typex.ToAnyE[V](val)
}

// GetMulti gets multiple fields from the hash
func (h *HashMap[K, V]) GetMulti(ctx context.Context, fields []K) (map[K]V, error) {
	if len(fields) == 0 {
		return make(map[K]V), nil
	}

	fieldStrs := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldStrs = append(fieldStrs, typex.ToString(field))
	}

	vals, err := h.Cli.HMGet(ctx, h.Key, fieldStrs...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[K]V, len(fields))
	for i, val := range vals {
		if val != nil {
			v, err := typex.ToAnyE[V](val.(string))
			if err != nil {
				return nil, err
			}
			result[fields[i]] = v
		}
	}

	return result, nil
}

// GetAll gets all fields and values from the hash
func (h *HashMap[K, V]) GetAll(ctx context.Context) (map[K]V, error) {
	vals, err := h.Cli.HGetAll(ctx, h.Key).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[K]V, len(vals))
	for k, v := range vals {
		key, err := typex.ToAnyE[K](k)
		if err != nil {
			return nil, err
		}
		val, err := typex.ToAnyE[V](v)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}

	return result, nil
}

// Delete deletes fields from the hash
func (h *HashMap[K, V]) Delete(ctx context.Context, fields ...K) error {
	if len(fields) == 0 {
		return nil
	}

	fieldStrs := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldStrs = append(fieldStrs, typex.ToString(field))
	}

	return h.Cli.HDel(ctx, h.Key, fieldStrs...).Err()
}

// Exists checks if a field exists in the hash
func (h *HashMap[K, V]) Exists(ctx context.Context, field K) (bool, error) {
	return h.Cli.HExists(ctx, h.Key, typex.ToString(field)).Result()
}

// Len returns the number of fields in the hash
func (h *HashMap[K, V]) Len(ctx context.Context) (int64, error) {
	return h.Cli.HLen(ctx, h.Key).Result()
}

// Keys returns all field names in the hash
func (h *HashMap[K, V]) Keys(ctx context.Context) ([]K, error) {
	keys, err := h.Cli.HKeys(ctx, h.Key).Result()
	if err != nil {
		return nil, err
	}

	result := make([]K, 0, len(keys))
	for _, k := range keys {
		key, err := typex.ToAnyE[K](k)
		if err != nil {
			return nil, err
		}
		result = append(result, key)
	}

	return result, nil
}

// Values returns all values in the hash
func (h *HashMap[K, V]) Values(ctx context.Context) ([]V, error) {
	vals, err := h.Cli.HVals(ctx, h.Key).Result()
	if err != nil {
		return nil, err
	}

	result := make([]V, 0, len(vals))
	for _, v := range vals {
		val, err := typex.ToAnyE[V](v)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}

	return result, nil
}

// Incr increments the integer value of a field by the given amount
func (h *HashMap[K, V]) Incr(ctx context.Context, field K, increment int64, expire time.Duration) (int64, error) {
	pipe := h.Cli.Pipeline()
	incrCmd := pipe.HIncrBy(ctx, h.Key, typex.ToString(field), increment)
	if expire > 0 {
		pipe.Expire(ctx, h.Key, expire)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}

// IncrFloat increments the float value of a field by the given amount
func (h *HashMap[K, V]) IncrFloat(ctx context.Context, field K, increment float64, expire time.Duration) (float64, error) {
	pipe := h.Cli.Pipeline()
	incrCmd := pipe.HIncrByFloat(ctx, h.Key, typex.ToString(field), increment)
	if expire > 0 {
		pipe.Expire(ctx, h.Key, expire)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}
