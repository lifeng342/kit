package repox

import "C"
import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormRepo GORM 通用仓库实现（基于 gorm.G 泛型 API）
type GormRepo[T any] struct {
	db *gorm.DB
}

// 确保 GormRepo 实现了 Repo 接口
var _ Repo[any] = (*GormRepo[any])(nil)

// NewGormRepo 创建 GORM 仓库
func NewGormRepo[T any](db *gorm.DB) *GormRepo[T] {
	return &GormRepo[T]{db: db}
}

// getDB 获取实际使用的 DB（优先从 context 获取事务 DB）
func (r *GormRepo[T]) getDB(ctx context.Context) *gorm.DB {
	if tx := GetGormTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Transaction 执行事务
// 使用示例：
//
//	err := repo.Transaction(ctx, func(ctx context.Context) error {
//	    // 直接使用原有的 repo，会自动使用事务 DB
//	    if err := userRepo.Create(ctx, user); err != nil {
//	        return err
//	    }
//	    return orderRepo.Create(ctx, order)
//	})
func (r *GormRepo[T]) Transaction(ctx context.Context, fn TxFunc) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := WithGormTx(ctx, tx)
		return fn(txCtx)
	})
}

// Context keys for transaction
type contextKey string

const gormTxKey contextKey = "gorm_tx"

// WithGormTx 将 GORM 事务 DB 注入到 context
func WithGormTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, gormTxKey, tx)
}

// GetGormTx 从 context 获取 GORM 事务 DB，如果不存在则返回 nil
func GetGormTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(gormTxKey).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// Native 返回底层 *gorm.DB
func (r *GormRepo[T]) Native() *gorm.DB {
	return r.db
}

// Create 创建单条记录
func (r *GormRepo[T]) Create(ctx context.Context, entity *T) error {
	return wrapError(gorm.G[T](r.getDB(ctx)).Create(ctx, entity))
}

// CreateMany 批量创建记录
func (r *GormRepo[T]) CreateMany(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	v := FromPtrSlice(entities)
	return wrapError(gorm.G[T](r.getDB(ctx)).CreateInBatches(ctx, &v, 10))
}

// FindOne 查询单条记录
func (r *GormRepo[T]) FindOne(ctx context.Context, filter any, opts ...IList[FindOptions]) (*T, error) {
	g := gorm.G[T](r.getDB(ctx))
	o := NewOptions(opts...)

	// 应用条件
	chain := r.applyFilterToChain(g, filter)
	chain = r.applyFindOptionsToChain(chain, o)

	result, err := chain.First(ctx)
	if err != nil {
		return nil, wrapError(err)
	}
	return &result, nil
}

// Find 查询多条记录
func (r *GormRepo[T]) Find(ctx context.Context, filter any, opts ...IList[FindOptions]) ([]*T, error) {
	g := gorm.G[T](r.getDB(ctx))
	o := NewOptions(opts...)

	chain := r.applyFilterToChain(g, filter)
	chain = r.applyFindOptionsToChain(chain, o)

	results, err := chain.Find(ctx)
	return ToPtrSlice(results), wrapError(err)
}

// Count 统计记录数
func (r *GormRepo[T]) Count(ctx context.Context, filter any, opts ...IList[FindOptions]) (int64, error) {
	g := gorm.G[T](r.getDB(ctx))
	chain := r.applyFilterToChain(g, filter)
	return chain.Count(ctx, "id")
}

// applyFilterToChain 应用过滤条件到链式调用
func (r *GormRepo[T]) applyFilterToChain(g gorm.Interface[T], filter any) gorm.ChainInterface[T] {
	return g.Where(filter)
}

// applyFindOptionsToChain 应用查询选项到链式调用
func (r *GormRepo[T]) applyFindOptionsToChain(chain gorm.ChainInterface[T], o *FindOptions) gorm.ChainInterface[T] {
	if len(o.ReturnFields) > 0 {
		chain = chain.Select(o.ReturnFields[0], ToAnySlice(o.ReturnFields[1:])...)
	}
	if o.Skip > 0 {
		chain = chain.Offset(int(o.Skip))
	}
	if o.Limit > 0 {
		chain = chain.Limit(int(o.Limit))
	}
	if o.Sort != nil {
		chain = chain.Order(o.Sort.ToSqlStr())
	}
	return chain
}

// Update 更新整个实体（通过主键）
func (r *GormRepo[T]) Update(ctx context.Context, entity *T) error {
	id, ok := getId(entity)
	if ok {
		_, err := gorm.G[T](r.getDB(ctx)).Where("id = ?", id).Updates(ctx, *entity)
		return wrapError(err)
	}

	return wrapError(r.getDB(ctx).WithContext(ctx).Save(entity).Error)
}

func (r *GormRepo[T]) Incr(ctx context.Context, filter any, incr map[string]int, opts ...IList[UpdateOptions]) error {
	var t T
	chain := r.getDB(ctx).WithContext(ctx).Model(t).Where(filter).Updates(r.incrToUpdate(incr))
	if chain.Error != nil {
		return wrapError(chain.Error)
	}
	return nil
}

// UpdateOne 更新单条记录
func (r *GormRepo[T]) UpdateOne(ctx context.Context, filter any, update map[string]any, opts ...IList[UpdateOptions]) (*UpdateResult, error) {
	//g := r.buildUpdateG(opts...)
	//chain := r.applyFilterToChain(g, filter)
	var t T
	chain := r.getDB(ctx).WithContext(ctx).Model(t).Clauses().Where(filter).Updates(update)
	if chain.Error != nil {
		return nil, wrapError(chain.Error)
	}
	return &UpdateResult{UpdateCount: chain.RowsAffected}, wrapError(nil)
}

// UpdateMany 更新多条记录
func (r *GormRepo[T]) UpdateMany(ctx context.Context, filter any, update map[string]any, opts ...IList[UpdateOptions]) (*UpdateResult, error) {
	//g := r.buildUpdateG(opts...)
	//chain := r.applyFilterToChain(g, filter)

	var t T
	chain := r.getDB(ctx).WithContext(ctx).Model(t).Where(filter).Updates(update)
	if chain.Error != nil {
		return nil, wrapError(chain.Error)
	}
	return &UpdateResult{UpdateCount: chain.RowsAffected}, wrapError(nil)
}

// UpsertOne 插入或更新单条记录
func (r *GormRepo[T]) UpsertOne(ctx context.Context, create T, opt UpsertOptions) error {
	columns := make([]clause.Column, 0, len(opt.ConflictKvs))
	for k := range opt.ConflictKvs {
		columns = append(columns, clause.Column{Name: k})
	}

	doUpdates := make(map[string]any)
	for k, v := range opt.Set {
		doUpdates[k] = v
	}
	for k, v := range opt.Inc {
		doUpdates[k] = gorm.Expr(k+" + ?", v)
	}

	return wrapError(gorm.G[T](r.getDB(ctx), clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.Assignments(doUpdates),
	}).Create(ctx, &create))
}

// buildUpdateG 构建带更新选项的泛型实例
func (r *GormRepo[T]) buildUpdateG(ctx context.Context, opts ...IList[UpdateOptions]) gorm.Interface[T] {
	_ = NewOptions(opts...)
	var clauses []clause.Expression

	//if o.Upsert {
	//	conflictCols := o.OnConflictColumn
	//	if len(conflictCols) == 0 {
	//		conflictCols = []string{"id"}
	//	}
	//	columns := make([]clause.Column, len(conflictCols))
	//	for i, col := range conflictCols {
	//		columns[i] = clause.Column{Name: col}
	//	}
	//	clauses = append(clauses, clause.OnConflict{
	//		Columns:   columns,
	//		UpdateAll: true,
	//	})
	//}

	return gorm.G[T](r.getDB(ctx), clauses...)
}

// DeleteOne 删除单条记录
func (r *GormRepo[T]) DeleteOne(ctx context.Context, filter any) (*DeleteResult, error) {
	g := gorm.G[T](r.getDB(ctx))
	chain := r.applyFilterToChain(g, filter)
	rowsAffected, err := chain.Limit(1).Delete(ctx)
	return &DeleteResult{DeleteCount: int64(rowsAffected)}, err
}

// DeleteMany 删除多条记录
func (r *GormRepo[T]) DeleteMany(ctx context.Context, filter any) (*DeleteResult, error) {
	g := gorm.G[T](r.getDB(ctx))
	chain := r.applyFilterToChain(g, filter)
	rowsAffected, err := chain.Delete(ctx)
	return &DeleteResult{DeleteCount: int64(rowsAffected)}, err
}

func (r *GormRepo[T]) incrToUpdate(incr map[string]int) map[string]any {
	ret := make(map[string]any)
	for k, v := range incr {
		ret[k] = gorm.Expr(k+" + ?", v)
	}
	return ret
}

func AsGormRepo[T any](repo Repo[T]) (*GormRepo[T], bool) {
	if gr, ok := repo.(*GormRepo[T]); ok {
		return gr, true
	}
	return nil, false
}

func GetNativeDB[T any](repo Repo[T]) (*gorm.DB, bool) {
	if gr, ok := AsGormRepo(repo); ok {
		return gr.Native(), true
	}
	return nil, false
}
