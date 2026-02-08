package redisx

import (
	"context"
	"time"

	"github.com/mbeoliero/kit/utils/typex"
	"github.com/redis/go-redis/v9"
)

// Element z set element
type Element[T any] struct {
	Member T     `json:"member"`
	Score  int64 `json:"score"`
}

type ElementList[T any] []Element[T]

func (e ElementList[T]) Members() []T {
	var members []T
	for _, v := range e {
		members = append(members, v.Member)
	}
	return members
}

type ZQueue[T any] struct {
	Key  string
	Cli  redis.UniversalClient
	Desc bool // true for descending order, false for ascending order
}

// redisZToElement from redis.Z to Element
func redisZToElement[T any](z redis.Z) Element[T] {
	return Element[T]{
		Member: typex.ToAny[T](z.Member.(string)),
		Score:  int64(z.Score),
	}
}

// redisZToElements from redis.Z slice to Element slice
func redisZToElements[T any](zs []redis.Z) []Element[T] {
	elements := make([]Element[T], 0, len(zs))
	for _, z := range zs {
		elements = append(elements, redisZToElement[T](z))
	}
	return elements
}

// Add adds an element to the sorted set with the given score
func (q *ZQueue[T]) Add(ctx context.Context, member T, score int64, expire time.Duration) error {
	pipe := q.Cli.Pipeline()
	pipe.ZAdd(ctx, q.Key, redis.Z{
		Score:  float64(score),
		Member: typex.ToString(member),
	})
	if expire > 0 {
		pipe.Expire(ctx, q.Key, expire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// AddMulti adds multiple elements to the sorted set
func (q *ZQueue[T]) AddMulti(ctx context.Context, elements []Element[T], expire time.Duration) error {
	members := make([]redis.Z, 0, len(elements))
	for _, elem := range elements {
		members = append(members, redis.Z{
			Score:  float64(elem.Score),
			Member: typex.ToString(elem.Member),
		})
	}
	pipe := q.Cli.Pipeline()
	pipe.ZAdd(ctx, q.Key, members...)
	if expire > 0 {
		pipe.Expire(ctx, q.Key, expire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Remove removes an element from the sorted set
func (q *ZQueue[T]) Remove(ctx context.Context, member T) error {
	return q.Cli.ZRem(ctx, q.Key, typex.ToString(member)).Err()
}

// RemoveMulti removes multiple elements from the sorted set
func (q *ZQueue[T]) RemoveMulti(ctx context.Context, members []T) error {
	memberStrs := make([]interface{}, 0, len(members))
	for _, member := range members {
		memberStrs = append(memberStrs, typex.ToString(member))
	}
	return q.Cli.ZRem(ctx, q.Key, memberStrs...).Err()
}

// RangeByScore returns elements with scores between min and max
// Respects the Desc field in ZQueue
func (q *ZQueue[T]) RangeByScore(ctx context.Context, minScore, maxScore int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, minScore, maxScore, 0, -1, q.Desc)
}

// RangeByScoreWithLimit returns elements with scores between min and max with pagination
// Respects the Desc field in ZQueue
func (q *ZQueue[T]) RangeByScoreWithLimit(ctx context.Context, minScore, maxScore int64, offset, count int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, minScore, maxScore, offset, count, q.Desc)
}

// RangeFromScore returns elements with scores >= minScore
// Respects the Desc field in ZQueue
func (q *ZQueue[T]) RangeFromScore(ctx context.Context, minScore int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, minScore, -1, 0, -1, q.Desc)
}

// RangeToScore returns elements with scores <= maxScore
// Respects the Desc field in ZQueue
func (q *ZQueue[T]) RangeToScore(ctx context.Context, maxScore int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, -1, maxScore, 0, -1, q.Desc)
}

// RangeByScoreRev returns elements with scores between min and max in reversed order
// Reverses the Desc field in ZQueue
func (q *ZQueue[T]) RangeByScoreRev(ctx context.Context, minScore, maxScore int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, minScore, maxScore, 0, -1, !q.Desc)
}

// RangeByScoreWithLimitRev returns elements with scores between min and max with pagination in reversed order
// Reverses the Desc field in ZQueue
func (q *ZQueue[T]) RangeByScoreWithLimitRev(ctx context.Context, minScore, maxScore int64, offset, count int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, minScore, maxScore, offset, count, !q.Desc)
}

// RangeFromScoreRev returns elements with scores >= minScore in reversed order
// Reverses the Desc field in ZQueue
func (q *ZQueue[T]) RangeFromScoreRev(ctx context.Context, minScore int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, minScore, -1, 0, -1, !q.Desc)
}

// RangeToScoreRev returns elements with scores <= maxScore in reversed order
// Reverses the Desc field in ZQueue
func (q *ZQueue[T]) RangeToScoreRev(ctx context.Context, maxScore int64) ([]Element[T], error) {
	return q.rangeByScoreInternal(ctx, -1, maxScore, 0, -1, !q.Desc)
}

// rangeByScoreInternal internal method to handle all range by score queries
func (q *ZQueue[T]) rangeByScoreInternal(ctx context.Context, minScore, maxScore int64, offset, count int64, desc bool) ([]Element[T], error) {
	minS := "-inf"
	if minScore != -1 {
		minS = typex.ToString(minScore)
	}

	maxS := "+inf"
	if maxScore != -1 {
		maxS = typex.ToString(maxScore)
	}

	var zs []redis.Z
	var err error

	if desc {
		zs, err = q.Cli.ZRevRangeByScoreWithScores(ctx, q.Key, &redis.ZRangeBy{
			Min:    minS,
			Max:    maxS,
			Offset: offset,
			Count:  count,
		}).Result()
	} else {
		zs, err = q.Cli.ZRangeByScoreWithScores(ctx, q.Key, &redis.ZRangeBy{
			Min:    minS,
			Max:    maxS,
			Offset: offset,
			Count:  count,
		}).Result()
	}

	if err != nil {
		return nil, err
	}
	return redisZToElements[T](zs), nil
}

// PopMin removes and returns the element with the lowest score
func (q *ZQueue[T]) PopMin(ctx context.Context) (*Element[T], error) {
	zs, err := q.Cli.ZPopMin(ctx, q.Key, 1).Result()
	if err != nil {
		return nil, err
	}
	if len(zs) == 0 {
		return nil, nil
	}
	elem := redisZToElement[T](zs[0])
	return &elem, nil
}

// PopMax removes and returns the element with the highest score
func (q *ZQueue[T]) PopMax(ctx context.Context) (*Element[T], error) {
	zs, err := q.Cli.ZPopMax(ctx, q.Key, 1).Result()
	if err != nil {
		return nil, err
	}
	if len(zs) == 0 {
		return nil, nil
	}
	elem := redisZToElement[T](zs[0])
	return &elem, nil
}

// PopMinMulti removes and returns multiple elements with the lowest scores
func (q *ZQueue[T]) PopMinMulti(ctx context.Context, count int64) ([]Element[T], error) {
	zs, err := q.Cli.ZPopMin(ctx, q.Key, count).Result()
	if err != nil {
		return nil, err
	}
	return redisZToElements[T](zs), nil
}

// PopMaxMulti removes and returns multiple elements with the highest scores
func (q *ZQueue[T]) PopMaxMulti(ctx context.Context, count int64) ([]Element[T], error) {
	zs, err := q.Cli.ZPopMax(ctx, q.Key, count).Result()
	if err != nil {
		return nil, err
	}
	return redisZToElements[T](zs), nil
}

// RemoveRangeByScore removes elements with scores between min and max
// Use "-inf" for min or "+inf" for max to represent infinity
func (q *ZQueue[T]) RemoveRangeByScore(ctx context.Context, min, max string) (int64, error) {
	return q.Cli.ZRemRangeByScore(ctx, q.Key, min, max).Result()
}

// Count returns the number of elements in the sorted set
func (q *ZQueue[T]) Count(ctx context.Context) (int64, error) {
	return q.Cli.ZCard(ctx, q.Key).Result()
}

// CountByScore returns the number of elements with scores between min and max
// Use "-inf" for min or "+inf" for max to represent infinity
func (q *ZQueue[T]) CountByScore(ctx context.Context, min, max string) (int64, error) {
	return q.Cli.ZCount(ctx, q.Key, min, max).Result()
}

// Score returns the score of a member
func (q *ZQueue[T]) Score(ctx context.Context, member T) (int64, error) {
	score, err := q.Cli.ZScore(ctx, q.Key, typex.ToString(member)).Result()
	if err != nil {
		return 0, err
	}
	return int64(score), nil
}
