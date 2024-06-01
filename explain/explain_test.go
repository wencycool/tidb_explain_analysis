package explain

import (
	"testing"
)

func TestNewPlanTree_FormatTypePlanBriefText(t *testing.T) {
	planText := `
	+-------------------------------------------------------------+-------------+-----------+----------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
	| id                                                          | estRows     | task      | access object  | operator info                                                                                                                                                                   |
	+-------------------------------------------------------------+-------------+-----------+----------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
	| Sort_30                                                     | 736.86      | root      |                | Column#62                                                                                                                                                                       |
	| └─Projection_32                                             | 736.86      | root      |                | Column#62, div(Column#64, Column#65)->Column#66                                                                                                                                 |
	|   └─HashAgg_33                                              | 736.86      | root      |                | group by:Column#73, funcs:sum(Column#70)->Column#64, funcs:sum(Column#71)->Column#65, funcs:firstrow(Column#72)->Column#62                                                      |
	|     └─Projection_120                                        | 116460.04   | root      |                | case(eq(tpch10.nation.n_name, BRAZIL), Column#63, 0.0000)->Column#70, Column#63->Column#71, Column#62->Column#72, Column#62->Column#73                                          |
	|       └─Projection_34                                       | 116460.04   | root      |                | extract(YEAR, tpch10.orders.o_orderdate)->Column#62, mul(tpch10.lineitem.l_extendedprice, minus(1, tpch10.lineitem.l_discount))->Column#63, tpch10.nation.n_name                |
	|         └─Projection_35                                     | 116460.04   | root      |                | tpch10.lineitem.l_extendedprice, tpch10.lineitem.l_discount, tpch10.orders.o_orderdate, tpch10.nation.n_name                                                                    |
	|           └─HashJoin_45                                     | 116460.04   | root      |                | inner join, equal:[eq(tpch10.supplier.s_nationkey, tpch10.nation.n_nationkey)]                                                                                                  |
	|             ├─TableReader_118(Build)                        | 25.00       | root      |                | data:TableFullScan_117                                                                                                                                                          |
	|             │ └─TableFullScan_117                           | 25.00       | cop[tikv] | table:n2       | keep order:false                                                                                                                                                                |
	|             └─HashJoin_56(Probe)                            | 116460.04   | root      |                | inner join, equal:[eq(tpch10.lineitem.l_suppkey, tpch10.supplier.s_suppkey)]                                                                                                    |
	|               ├─TableReader_116(Build)                      | 100000.00   | root      |                | data:TableFullScan_115                                                                                                                                                          |
	|               │ └─TableFullScan_115                         | 100000.00   | cop[tikv] | table:supplier | keep order:false                                                                                                                                                                |
	|               └─HashJoin_69(Probe)                          | 115714.70   | root      |                | inner join, equal:[eq(tpch10.lineitem.l_partkey, tpch10.part.p_partkey)]                                                                                                        |
	|                 ├─TableReader_114(Build)                    | 13015.52    | root      |                | data:Selection_113                                                                                                                                                              |
	|                 │ └─Selection_113                           | 13015.52    | cop[tikv] |                | eq(tpch10.part.p_type, "ECONOMY ANODIZED STEEL")                                                                                                                                |
	|                 │   └─TableFullScan_112                     | 2000000.00  | cop[tikv] | table:part     | keep order:false                                                                                                                                                                |
	|                 └─IndexHashJoin_78(Probe)                   | 17732099.08 | root      |                | inner join, inner:TableReader_73, outer key:tpch10.orders.o_orderkey, inner key:tpch10.lineitem.l_orderkey, equal cond:eq(tpch10.orders.o_orderkey, tpch10.lineitem.l_orderkey) |
	|                   ├─HashJoin_84(Build)                      | 4456928.25  | root      |                | inner join, equal:[eq(tpch10.customer.c_custkey, tpch10.orders.o_custkey)]                                                                                                      |
	|                   │ ├─HashJoin_86(Build)                    | 300000.00   | root      |                | inner join, equal:[eq(tpch10.nation.n_nationkey, tpch10.customer.c_nationkey)]                                                                                                  |
	|                   │ │ ├─HashJoin_99(Build)                  | 5.00        | root      |                | inner join, equal:[eq(tpch10.region.r_regionkey, tpch10.nation.n_regionkey)]                                                                                                    |
	|                   │ │ │ ├─TableReader_104(Build)            | 1.00        | root      |                | data:Selection_103                                                                                                                                                              |
	|                   │ │ │ │ └─Selection_103                   | 1.00        | cop[tikv] |                | eq(tpch10.region.r_name, "AMERICA")                                                                                                                                             |
	|                   │ │ │ │   └─TableFullScan_102             | 5.00        | cop[tikv] | table:region   | keep order:false                                                                                                                                                                |
	|                   │ │ │ └─TableReader_101(Probe)            | 25.00       | root      |                | data:TableFullScan_100                                                                                                                                                          |
	|                   │ │ │   └─TableFullScan_100               | 25.00       | cop[tikv] | table:n1       | keep order:false                                                                                                                                                                |
	|                   │ │ └─TableReader_106(Probe)              | 1500000.00  | root      |                | data:TableFullScan_105                                                                                                                                                          |
	|                   │ │   └─TableFullScan_105                 | 1500000.00  | cop[tikv] | table:customer | keep order:false                                                                                                                                                                |
	|                   │ └─TableReader_109(Probe)                | 4593898.95  | root      |                | data:Selection_108                                                                                                                                                              |
	|                   │   └─Selection_108                       | 4593898.95  | cop[tikv] |                | ge(tpch10.orders.o_orderdate, 1995-01-01), le(tpch10.orders.o_orderdate, 1996-12-31)                                                                                            |
	|                   │     └─TableFullScan_107                 | 15000000.00 | cop[tikv] | table:orders   | keep order:false                                                                                                                                                                |
	|                   └─TableReader_73(Probe)                   | 4456928.25  | root      |                | data:TableRangeScan_72                                                                                                                                                          |
	|                     └─TableRangeScan_72                     | 4456928.25  | cop[tikv] | table:lineitem | range: decided by [eq(tpch10.lineitem.l_orderkey, tpch10.orders.o_orderkey)], keep order:false                                                                                  |
	+-------------------------------------------------------------+-------------+-----------+----------------+---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+

	`
	rawPlan, err := GetRawPlanFromText(planText, FormatTypePlanBriefText)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	planTree, err := NewPlanTree(rawPlan)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if planTree == nil {
		t.Fatalf("Expected a plan tree, but got nil")
	}
	planTree.Traverse()

}

func TestNewPlanTree_FormatTypePlanBriefText1(t *testing.T) {
	planText := `
	+------------------------------+---------+-----------+---------------+----------------------------------------------+
	| id                           | estRows | task      | access object | operator info                                |
	+------------------------------+---------+-----------+---------------+----------------------------------------------+
	| HashJoin_22                  | 2.00    | root      |               | inner join, equal:[eq(test.a.c1, test.b.c1)] |
	| ├─TableReader_25(Build)      | 2.00    | root      |               | data:Selection_24                            |
	| │ └─Selection_24             | 2.00    | cop[tikv] |               | not(isnull(test.a.c1))                       |
	| │   └─TableFullScan_23       | 2.00    | cop[tikv] | table:a       | keep order:false                             |
	| └─TableReader_28(Probe)      | 6.00    | root      |               | data:Selection_27                            |
	|   └─Selection_27             | 6.00    | cop[tikv] |               | not(isnull(test.b.c1))                       |
	|     └─TableFullScan_26       | 6.00    | cop[tikv] | table:b       | keep order:false                             |
	+------------------------------+---------+-----------+---------------+----------------------------------------------+
	7 rows in set (0.00 sec)
	`
	rawPlan, err := GetRawPlanFromText(planText, FormatTypePlanBriefText)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	planTree, err := NewPlanTree(rawPlan)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if planTree == nil {
		t.Fatalf("Expected a plan tree, but got nil")
	}
	planTree.Traverse()

}

func TestNewPlanTree_FormatTypeAnalyzeVerboseText(t *testing.T) {
	planText := `
	+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
	| 
	| id                    | estRows  | estCost   | actRows | task | access object            | execution info                         | operator info                            | memory  | disk  |
	| Projection_15         | 1.00     | 499006.50 | 1       | root |                          | time:342.9µs, loops:2, Concurrency:OFF | Column#91                                | 1.24 KB | N/A   |
	| └─TopN_7              | 1.00     | 499006.40 | 1       | root |                          | time:339.9µs, loops:2                  | Column#103:desc, offset:0, count:1       | 2.16 KB | N/A   |
	|   └─Projection_16     | 10000.00 | 100798.00 | 5       | root |                          | time:315.9µs, loops:2, Concurrency:5   | Column#91, length(Column#91)->Column#103 | 9.19 KB | N/A   |
	|     └─MemTableScan_11 | 10000.00 | 0.00      | 5       | root | table:STATEMENTS_SUMMARY | time:190.1µs, loops:2                  |                                          | N/A     | N/A   |
					 |
	+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
	1 row in set (0.00 sec)
	`
	rawPlan, err := GetRawPlanFromText(planText, FormatTypeAnalyzeVerboseText)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	planTree, err := NewPlanTree(rawPlan)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if planTree == nil {
		t.Fatalf("Expected a plan tree, but got nil")
	}
	planTree.Traverse()

}
