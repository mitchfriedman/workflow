package tracing

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/mitchfriedman/workflow/lib/logging"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"

	"github.com/jinzhu/gorm"
)

type contextKey struct{}

var methodNameKey = contextKey{}

const (
	parentSpanGormKey = "tracingParentSpan"
	spanGormKey       = "tracingSpan"
)

// AddGormCallbacks instruments the instance of db with the callbacks. It should be used with
// NewDBSpan one each call to a db operation.
func AddGormCallbacks(db *gorm.DB, logger logging.StructuredLogger) {
	callbacks := newCallbacks(logger)
	registerCallbacks(db, "create", callbacks)
	registerCallbacks(db, "query", callbacks)
	registerCallbacks(db, "update", callbacks)
	registerCallbacks(db, "delete", callbacks)
	registerCallbacks(db, "row_query", callbacks)
}

type callbacks struct {
	logger logging.StructuredLogger
}

func newCallbacks(logger logging.StructuredLogger) *callbacks {
	return &callbacks{logger: logger}
}

func (c *callbacks) beforeCreate(scope *gorm.Scope)   { c.before(scope) }
func (c *callbacks) afterCreate(scope *gorm.Scope)    { c.after(scope, "INSERT") }
func (c *callbacks) beforeQuery(scope *gorm.Scope)    { c.before(scope) }
func (c *callbacks) afterQuery(scope *gorm.Scope)     { c.after(scope, "SELECT") }
func (c *callbacks) beforeUpdate(scope *gorm.Scope)   { c.before(scope) }
func (c *callbacks) afterUpdate(scope *gorm.Scope)    { c.after(scope, "UPDATE") }
func (c *callbacks) beforeDelete(scope *gorm.Scope)   { c.before(scope) }
func (c *callbacks) afterDelete(scope *gorm.Scope)    { c.after(scope, "DELETE") }
func (c *callbacks) beforeRowQuery(scope *gorm.Scope) { c.before(scope) }
func (c *callbacks) afterRowQuery(scope *gorm.Scope)  { c.after(scope, "") }

func (c *callbacks) after(scope *gorm.Scope, operation string) {
	val, ok := scope.Get(spanGormKey)
	if !ok {
		c.logger.Warnf("failed to fetch span from scope. Did `before` fail?")
		return
	}
	sp := val.(Span)
	if operation == "" {
		operation = strings.ToUpper(strings.Split(scope.SQL, " ")[0])
	}
	sp.SetTag(ext.Error, scope.HasError())
	sp.SetTag(ext.DBStatement, scope.SQL)
	sp.SetTag("db.table", scope.TableName())
	sp.SetTag("db.method", operation)
	sp.SetTag("db.err", scope.HasError())
	sp.SetTag("db.count", scope.DB().RowsAffected)
	sp.RecordError(scope.DB().Error)
	sp.Finish()
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("tracing:%v_before", name)
	afterName := fmt.Sprintf("tracing:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)
	// gorm does some magic, if you pass CallbackProcessor here - nothing works
	switch name {
	case "create":
		db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
	case "query":
		db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
	case "delete":
		db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
	case "row_query":
		db.Callback().RowQuery().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		db.Callback().RowQuery().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}

func (c *callbacks) before(scope *gorm.Scope) {
	val, ok := scope.Get(parentSpanGormKey)
	if !ok {
		c.logger.Warnf("failed to fetch parent span from gorm scope. Did you forget to call `NewDBSpan`?")
		return
	}

	parentSpan := val.(Span)

	// TODO: replace with context from span when this is supported.
	sp, _ := tracer.StartSpanFromContext(context.TODO(), parentSpan.operationName, tracer.SpanType("db"))
	sp.SetTag("sql.query", scope.SQL)
	sp.SetTag("database", scope.InstanceID())

	internalSpan := Span{span: sp, operationName: parentSpan.operationName}
	scope.Set(spanGormKey, internalSpan)
}
