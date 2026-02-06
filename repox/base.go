package repox

import (
	"context"
	"errors"
)

var DataNotFound = errors.New("data not found")

// TxFunc 事务函数类型
type TxFunc func(ctx context.Context) error

// FindOptions 存储查询配置
type FindOptions struct {
	ReturnFields []string
	Skip         int64
	Limit        int64
	Sort         *Sort
}

// FindOptionsBuilder 链式构建器
type FindOptionsBuilder struct {
	Opts []func(*FindOptions)
}

// Find 创建新的构建器
func Find() *FindOptionsBuilder {
	return &FindOptionsBuilder{}
}

// List 返回所有配置函数
func (f *FindOptionsBuilder) List() []func(*FindOptions) {
	return f.Opts
}

func (f *FindOptionsBuilder) SetReturnFields(fields ...string) *FindOptionsBuilder {
	f.Opts = append(f.Opts, func(opts *FindOptions) {
		opts.ReturnFields = fields
	})
	return f
}

func (f *FindOptionsBuilder) SetSkip(skip int64) *FindOptionsBuilder {
	f.Opts = append(f.Opts, func(opts *FindOptions) {
		opts.Skip = skip
	})
	return f
}

func (f *FindOptionsBuilder) SetLimit(limit int64) *FindOptionsBuilder {
	f.Opts = append(f.Opts, func(opts *FindOptions) {
		opts.Limit = limit
	})
	return f
}

func (f *FindOptionsBuilder) SetSort(sort *Sort) *FindOptionsBuilder {
	f.Opts = append(f.Opts, func(opts *FindOptions) {
		opts.Sort = sort
	})
	return f
}

// UpdateOptions 存储更新配置
type UpdateOptions struct {
}

// UpdateOptionsBuilder 链式构建器
type UpdateOptionsBuilder struct {
	Opts []func(*UpdateOptions)
}

// Update 创建新的构建器
func Update() *UpdateOptionsBuilder {
	return &UpdateOptionsBuilder{}
}

// List 返回所有配置函数
func (u *UpdateOptionsBuilder) List() []func(*UpdateOptions) {
	return u.Opts
}

type UpsertOptions struct {
	ConflictKvs map[string]any   // 冲突字段（唯一索引），用作filter
	Set         map[string]any   // 普通赋值
	Inc         map[string]int64 // 自增
}

type ICreator[T any] interface {
	Create(context.Context, *T) error
	CreateMany(context.Context, []*T) error
}

type IFinder[T any] interface {
	FindOne(ctx context.Context, filter any, opts ...IList[FindOptions]) (*T, error)
	Find(ctx context.Context, filter any, opts ...IList[FindOptions]) ([]*T, error)
	Count(ctx context.Context, filter any, opts ...IList[FindOptions]) (int64, error)
}

type UpdateResult struct {
	UpdateCount int64
}

type IUpdater[T any] interface {
	Update(context.Context, *T) error
	Incr(ctx context.Context, filter any, incr map[string]int, opts ...IList[UpdateOptions]) error
	UpdateOne(ctx context.Context, filter any, update map[string]any, opts ...IList[UpdateOptions]) (*UpdateResult, error)
	UpsertOne(ctx context.Context, create T, opt UpsertOptions) error
	UpdateMany(ctx context.Context, filter any, update map[string]any, opts ...IList[UpdateOptions]) (*UpdateResult, error)
}

type DeleteResult struct {
	DeleteCount int64
}

type IDeleter[T any] interface {
	DeleteOne(ctx context.Context, filter any) (*DeleteResult, error)
	DeleteMany(ctx context.Context, filter any) (*DeleteResult, error)
}

type ITransaction[T any] interface {
	Transaction(ctx context.Context, fn TxFunc) error
}

// INative 提供访问底层数据库连接的能力
// 对于 GORM 返回 *gorm.DB
// 对于 MongoDB 返回 *mongo.Collection
type INative[C any] interface {
	Native() C
}

type Repo[T any] interface {
	ICreator[T]
	IFinder[T]
	IUpdater[T]
	IDeleter[T]
	ITransaction[T]
}
