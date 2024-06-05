# tidb_explain_analysis

#### 介绍
用于解析 tidb的执行计划

调用方式
```go
// 1. 生成执行计划信息
planText := "explain analyze format='verbose' from select * from test where id = 1"
// 2. 解析执行计划，放到树形结构中
rawPlan, _ := GetRawPlanFromText(planText, FormatTypeAnalyzeVerboseText)
planTree, _ := NewPlanTree(rawPlan)
if planTree == nil {
t.Fatalf("Expected a plan tree, but got nil")
}
planTree.Traverse()
```