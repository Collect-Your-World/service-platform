package util

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	pagingUtil "backend/service-platform/app/pkg/util/paging"

	"github.com/mitchellh/mapstructure"
	"github.com/uptrace/bun"
)

type DBTable interface {
	Alias() string
}

var ErrorMissingAlias = fmt.Errorf("no alias for query containing relations. Please add Alias() method to the entity")

func StructToQueries(input any, alias string) (conds []string, args []any, err error) {
	output := make(map[string]any)
	if err = mapstructure.Decode(input, &output); err != nil {
		return nil, nil, err
	}
	if alias != "" {
		alias += "."
	}

	for key, value := range output {
		// check if value is array
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
			conds = append(conds, alias+key+" IN (?)")
			args = append(args, bun.In(value))
		} else {
			conds = append(conds, alias+key+" = ?")
			args = append(args, value)
		}
	}

	return conds, args, nil
}

func StructToConditions(input any, alias string) (condition string, args []any, err error) {
	conds, args, err := StructToQueries(input, alias)
	if err != nil {
		return "", nil, err
	}

	if len(conds) == 0 {
		return "", nil, nil
	}
	return strings.Join(conds, " AND "), args, nil
}

func FindOneEntityOrFailed[E DBTable, F any](
	ctx context.Context,
	db bun.IDB,
	filter F,
	selectors []string,
	relations ...string,
) (e E, err error) {
	entities, err := FindManyEntity[E, F](ctx, db, filter, selectors, pagingUtil.Page{Limit: 1}, relations...)
	if err != nil {
		return e, err
	}
	if len(entities) == 0 {
		return e, sql.ErrNoRows
	}
	return entities[0], nil
}

func FindManyEntity[E DBTable, F any](
	ctx context.Context,
	db bun.IDB,
	filter F,
	selectors []string,
	paging pagingUtil.Page,
	relations ...string,
) (entities []E, err error) {
	query := db.NewSelect().Model(&entities)
	query, err = BuildQueryConditions[E](query, filter, selectors, paging, relations...)
	if err != nil {
		return nil, err
	}
	err = query.Scan(ctx)
	return entities, SkipNotFound(err)
}

func BuildQueryConditions[E DBTable, F any](
	query *bun.SelectQuery,
	filter F,
	selectors []string,
	paging pagingUtil.Page,
	relations ...string,
) (*bun.SelectQuery, error) {
	var alias string
	if len(relations) > 0 {
		var e E
		alias = e.Alias()
	}
	paging.LoadDefault()
	condition, args, err := StructToConditions(filter, alias)
	if err != nil {
		return nil, err
	}
	if len(args) > 0 {
		query = query.Where(condition, args...)
	}
	for _, selector := range selectors {
		query = query.Column(selector)
	}
	for _, relation := range relations {
		query = query.Relation(relation)
	}

	query = query.Offset(paging.Offset).Limit(paging.Limit)
	if paging.OrderBy != "" {
		query = query.OrderExpr("? "+string(paging.SortBy), bun.Ident(paging.OrderBy))
	}
	return query, nil
}

func FindOneModelOrFailed[E any, F any](
	ctx context.Context,
	db bun.IDB,
	filter F,
	selectors ...string,
) (e E, err error) {
	entities, err := FindManyModel[E, F](ctx, db, filter, pagingUtil.Page{Limit: 1}, selectors...)
	if err != nil {
		return e, err
	}
	if len(entities) == 0 {
		return e, sql.ErrNoRows
	}
	return entities[0], nil
}

func FindManyModel[E any, F any](
	ctx context.Context,
	db bun.IDB,
	filter F,
	paging pagingUtil.Page,
	selectors ...string,
) (entities []E, err error) {
	paging.LoadDefault()
	condition, args, err := StructToConditions(filter, "")
	if err != nil {
		return nil, err
	}
	query := db.NewSelect().Model(&entities).Where(condition, args...)
	for _, selector := range selectors {
		query = query.Column(selector)
	}

	query = query.Offset(paging.Offset).Limit(paging.Limit)
	if paging.OrderBy != "" {
		query = query.OrderExpr("? ?", bun.Ident(paging.OrderBy), paging.SortBy)
	}
	err = query.Scan(ctx)
	return entities, SkipNotFound(err)
}

func UpdateBy[E any, F any, U any](
	ctx context.Context,
	db bun.IDB,
	filter F,
	data U,
) error {
	query := db.NewUpdate().Model((*E)(nil))
	setQueries, args, err := StructToQueries(data, "")
	if err != nil {
		return err
	}
	if len(setQueries) == 0 {
		return nil
	}
	for i, setq := range setQueries {
		query = query.Set(setq, args[i])
	}

	condition, args, err := StructToConditions(filter, "")
	if err != nil {
		return err
	}
	if len(args) > 0 {
		query = query.Where(condition, args...)
	}

	_, err = query.Exec(ctx)
	return err
}

func CheckExist[E any, F any](
	ctx context.Context,
	db bun.IDB,
	filter F,
) (bool, error) {
	condition, args, err := StructToConditions(filter, "")
	if err != nil {
		return false, err
	}

	return db.NewSelect().Model((*E)(nil)).Where(condition, args...).Exists(ctx)
}

func MatchAndUpdate[E any, U any](
	ctx context.Context,
	db bun.IDB,
	data []*U,
	filterFields []string,
	alias string,
) error {
	if len(data) == 0 {
		return nil
	}
	values := db.NewValues(&data)
	query := db.NewUpdate().
		With("_data", values).
		Model((*E)(nil)).
		TableExpr("_data")

	mData := make(map[string]any)
	if err := mapstructure.Decode(data[0], &mData); err != nil {
		return err
	}

	for key := range mData {
		query = query.Set("? = _data.?", bun.Ident(key), bun.Ident(key))
	}

	for _, field := range filterFields {
		query = query.Where("?.? = _data.?", bun.Ident(alias), bun.Ident(field), bun.Ident(field))
	}

	_, err := query.Exec(ctx)
	return err
}

func BuildUpdateQuery[E any, F any, U any](
	ctx context.Context,
	db bun.IDB,
	model *E,
	data U,
	filter F,
	alias string,
) (*bun.UpdateQuery, error) {
	query := db.NewUpdate().Model(model)

	setQueries, args, err := StructToQueries(data, alias)
	if err != nil {
		return nil, err
	}
	if len(setQueries) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	for i, setQuery := range setQueries {
		query = query.Set(setQuery, args[i])
	}

	condition, condArgs, err := StructToConditions(filter, alias)
	if err != nil {
		return nil, err
	}
	if condition != "" {
		query = query.Where(condition, condArgs...)
	}

	return query, nil
}

func FindManyEntityWithCount[E DBTable, F any](
	ctx context.Context,
	db bun.IDB,
	filter F,
	selectors []string,
	paging pagingUtil.Page,
	relations ...string,
) (entities []E, total int, err error) {
	query := db.NewSelect().Model(&entities)
	query, err = BuildQueryConditions[E](query, filter, selectors, paging, relations...)
	if err != nil {
		return nil, 0, err
	}
	total, err = query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, SkipNotFound(err)
	}

	return entities, total, nil
}
