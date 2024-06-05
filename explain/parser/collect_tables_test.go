package parser

import (
	"github.com/pingcap/tidb/pkg/parser"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
	"testing"
)

/*
mysql> explain select * from region r left join (select * from  customer a,tpch1.nation  where a.c_nationkey=n_nationkey) xx on r.R_REGIONKEY=xx.N_REGIONKEY;
+---------------------------------+-----------+-----------+---------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| id                              | estRows   | task      | access object | operator info                                                                                                                                                                                                                                                                                                   |
+---------------------------------+-----------+-----------+---------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| HashJoin_14                     | 37500.00  | root      |               | left outer join, equal:[eq(tpch1.region.r_regionkey, tpch1.nation.n_regionkey)]                                                                                                                                                                                                                                 |
| ├─TableReader_16(Build)         | 5.00      | root      |               | data:TableFullScan_15                                                                                                                                                                                                                                                                                           |
| │ └─TableFullScan_15            | 5.00      | cop[tikv] | table:r       | keep order:false, stats:pseudo                                                                                                                                                                                                                                                                                  |
| └─Projection_17(Probe)          | 150000.00 | root      |               | tpch1.customer.c_custkey, tpch1.customer.c_name, tpch1.customer.c_address, tpch1.customer.c_nationkey, tpch1.customer.c_phone, tpch1.customer.c_acctbal, tpch1.customer.c_mktsegment, tpch1.customer.c_comment, tpch1.nation.n_nationkey, tpch1.nation.n_name, tpch1.nation.n_regionkey, tpch1.nation.n_comment |
|   └─HashJoin_28                 | 150000.00 | root      |               | inner join, equal:[eq(tpch1.nation.n_nationkey, tpch1.customer.c_nationkey)]                                                                                                                                                                                                                                    |
|     ├─TableReader_32(Build)     | 25.00     | root      |               | data:TableFullScan_31                                                                                                                                                                                                                                                                                           |
|     │ └─TableFullScan_31        | 25.00     | cop[tikv] | table:nation  | keep order:false, stats:pseudo                                                                                                                                                                                                                                                                                  |
|     └─TableReader_30(Probe)     | 150000.00 | root      |               | data:TableFullScan_29                                                                                                                                                                                                                                                                                           |
|       └─TableFullScan_29        | 150000.00 | cop[tikv] | table:a       | keep order:false                                                                                                                                                                                                                                                                                                |
+---------------------------------+-----------+-----------+---------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
9 rows in set (0.00 sec)
*/

func TestCollectTableNamesVisitor_Enter(t *testing.T) {
	sel1 := "select * from region r left join (select * from  customer a,tpch1.nation  where a.c_nationkey=n_nationkey) xx on r.R_REGIONKEY=xx.N_REGIONKEY"
	p := parser.New()
	node, err := p.ParseOneStmt(sel1, "", "")
	if err != nil {
		panic(err)
	}
	c := new(CollectTableNamesVisitor)
	node.Accept(c)
	for _, tableName := range c.TableNames {
		t.Log(tableName)
	}
}
