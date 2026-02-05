package repox

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoRepo MongoDB 通用仓库实现
type MongoRepo[T any] struct {
	coll *mongo.Collection
}

// 确保 MongoRepo 实现了 Repo 接口
var _ Repo[any] = (*MongoRepo[any])(nil)

// NewMongoRepo 创建 MongoDB 仓库
func NewMongoRepo[T any](coll *mongo.Collection) *MongoRepo[T] {
	return &MongoRepo[T]{coll: coll}
}

// Native 返回底层 *mongo.Collection
func (r *MongoRepo[T]) Native() *mongo.Collection {
	return r.coll
}

// Create 创建单条记录
func (r *MongoRepo[T]) Create(ctx context.Context, entity *T) error {
	_, err := r.coll.InsertOne(ctx, entity)
	return wrapError(err)
}

// CreateMany 批量创建记录
func (r *MongoRepo[T]) CreateMany(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	_, err := r.coll.InsertMany(ctx, entities)
	return wrapError(err)
}

// FindOne 查询单条记录
func (r *MongoRepo[T]) FindOne(ctx context.Context, filter any, opts ...IList[FindOptions]) (*T, error) {
	o := NewOptions(opts...)
	findOpts := r.buildFindOneOptions(o)

	var result *T
	err := r.coll.FindOne(ctx, r.normalizeFilter(filter), findOpts).Decode(&result)
	if err != nil {
		return nil, wrapError(err)
	}
	return result, nil
}

// Find 查询多条记录
func (r *MongoRepo[T]) Find(ctx context.Context, filter any, opts ...IList[FindOptions]) ([]*T, error) {
	o := NewOptions(opts...)
	findOpts := r.buildFindOptions(o)

	cursor, err := r.coll.Find(ctx, r.normalizeFilter(filter), findOpts)
	if err != nil {
		return nil, wrapError(err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	var results []*T
	if err = cursor.All(ctx, &results); err != nil {
		return nil, wrapError(err)
	}
	return results, nil
}

// Count 统计记录数
func (r *MongoRepo[T]) Count(ctx context.Context, filter any, opts ...IList[FindOptions]) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, r.normalizeFilter(filter))
	return count, wrapError(err)
}

// Update 更新整个实体（通过 _id）
func (r *MongoRepo[T]) Update(ctx context.Context, entity *T) error {
	id, ok := getId(entity)
	if !ok {
		return errors.New("invalid entity")
	}
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": entity})
	return wrapError(err)
}

func (r *MongoRepo[T]) Incr(ctx context.Context, filter any, incr map[string]int, opts ...IList[UpdateOptions]) error {
	_ = NewOptions(opts...)
	updateOpts := options.UpdateOne()

	_, err := r.coll.UpdateOne(ctx, r.normalizeFilter(filter), r.incrToUpdate(incr), updateOpts)
	if err != nil {
		return wrapError(err)
	}
	return nil
}

// UpdateOne 更新单条记录
func (r *MongoRepo[T]) UpdateOne(ctx context.Context, filter any, update map[string]any, opts ...IList[UpdateOptions]) (*UpdateResult, error) {
	_ = NewOptions(opts...)
	updateOpts := options.UpdateOne()

	result, err := r.coll.UpdateOne(ctx, r.normalizeFilter(filter), r.mapToUpdate(update), updateOpts)
	if err != nil {
		return nil, wrapError(err)
	}
	return &UpdateResult{UpdateCount: result.ModifiedCount}, nil
}

// UpdateMany 更新多条记录
func (r *MongoRepo[T]) UpdateMany(ctx context.Context, filter any, update map[string]any, opts ...IList[UpdateOptions]) (*UpdateResult, error) {
	updateOpts := options.UpdateMany()

	result, err := r.coll.UpdateMany(ctx, r.normalizeFilter(filter), r.mapToUpdate(update), updateOpts)
	if err != nil {
		return nil, wrapError(err)
	}
	return &UpdateResult{UpdateCount: result.ModifiedCount}, nil
}

// UpsertOne 插入或更新单条记录
func (r *MongoRepo[T]) UpsertOne(ctx context.Context, create T, opt UpsertOptions) error {
	// 根据冲突字段构建 filter
	filter := bson.M{}
	for col, val := range opt.ConflictKvs {
		filter[col] = val
	}

	update := bson.M{}
	if len(opt.Set) > 0 {
		update["$set"] = opt.Set
	}
	if len(opt.Inc) > 0 {
		update["$inc"] = opt.Inc
	}
	// 如果没有指定 Set，则用整个 create 对象作为 $set
	if len(opt.Set) == 0 {
		update["$set"] = create
	}

	updateOpts := options.UpdateOne().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, updateOpts)
	return wrapError(err)
}

// DeleteOne 删除单条记录
func (r *MongoRepo[T]) DeleteOne(ctx context.Context, filter any) (*DeleteResult, error) {
	result, err := r.coll.DeleteOne(ctx, r.normalizeFilter(filter))
	if err != nil {
		return nil, wrapError(err)
	}
	return &DeleteResult{DeleteCount: result.DeletedCount}, nil
}

// DeleteMany 删除多条记录
func (r *MongoRepo[T]) DeleteMany(ctx context.Context, filter any) (*DeleteResult, error) {
	result, err := r.coll.DeleteMany(ctx, r.normalizeFilter(filter))
	if err != nil {
		return nil, wrapError(err)
	}
	return &DeleteResult{DeleteCount: result.DeletedCount}, nil
}

// buildFindOneOptions 构建 FindOne 选项
func (r *MongoRepo[T]) buildFindOneOptions(o *FindOptions) *options.FindOneOptionsBuilder {
	opts := options.FindOne()
	if len(o.ReturnFields) > 0 {
		projection := bson.M{}
		for _, f := range o.ReturnFields {
			projection[f] = 1
		}
		opts.SetProjection(projection)
	}
	if o.Skip > 0 {
		opts.SetSkip(o.Skip)
	}
	if o.Sort != nil {
		opts.SetSort(o.Sort.ToBson())
	}
	return opts
}

// buildFindOptions 构建 Find 选项
func (r *MongoRepo[T]) buildFindOptions(o *FindOptions) *options.FindOptionsBuilder {
	opts := options.Find()
	if len(o.ReturnFields) > 0 {
		projection := bson.M{}
		for _, f := range o.ReturnFields {
			projection[f] = 1
		}
		opts.SetProjection(projection)
	}
	if o.Skip > 0 {
		opts.SetSkip(o.Skip)
	}
	if o.Limit > 0 {
		opts.SetLimit(o.Limit)
	}
	if o.Sort != nil {
		opts.SetSort(o.Sort.ToBson())
	}
	return opts
}

func (r *MongoRepo[T]) incrToUpdate(incr map[string]int) bson.M {
	return bson.M{
		"$inc": incr,
	}
}

func (r *MongoRepo[T]) mapToUpdate(update map[string]any) bson.M {
	return bson.M{"$set": update}
}

// normalizeFilter 规范化过滤条件
func (r *MongoRepo[T]) normalizeFilter(filter any) any {
	if filter == nil {
		return bson.M{}
	}
	return filter
}

// getId 从实体中获取 _id 字段
func getId(entity any) (any, bool) {
	if e, ok := entity.(interface{ GetId() any }); ok {
		return e.GetId(), true
	}

	return nil, false
}

func AsMongoRepo[T any](repo Repo[T]) (*MongoRepo[T], bool) {
	if mr, ok := repo.(*MongoRepo[T]); ok {
		return mr, true
	}
	return nil, false
}

func GetNativeCollection[T any](repo Repo[T]) (*mongo.Collection, bool) {
	if mr, ok := AsMongoRepo(repo); ok {
		return mr.Native(), true
	}
	return nil, false
}
