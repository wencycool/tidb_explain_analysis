package parser

import "github.com/pingcap/tidb/pkg/parser/ast"

// 获取一条语句中的所有表名
// 实现了Visitor接口

type CollectTableNamesVisitor struct {
	defaultSchema string
	TableNames    []TableName
}

func (c *CollectTableNamesVisitor) Enter(n ast.Node) (node ast.Node, skipChildren bool) {
	switch x := n.(type) {
	case *ast.TableSource:
		switch y := x.Source.(type) {
		case *ast.TableName:
			tableName := TableName{
				Schema: y.Schema.O,
				Name:   y.Name.O,
				Alias:  x.AsName.O,
			}
			if tableName.Schema == "" {
				tableName.Schema = c.defaultSchema
			}
			c.TableNames = append(c.TableNames, tableName)
		}
	}
	return n, false

}
func (c *CollectTableNamesVisitor) Leave(n ast.Node) (node ast.Node, ok bool) {
	return n, true
}

type TableName struct {
	Schema string
	Name   string
	Alias  string
}
