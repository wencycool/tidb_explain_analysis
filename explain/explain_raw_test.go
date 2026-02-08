package explain

import (
	"reflect"
	"testing"
)

func TestGetHeaderColsPositionReturnsCorrectPositions(t *testing.T) {
	planText := `
+-----------------------+------------+-----------+----------------+----------------------+
| id                    | estRows    | task      | access object  | operator info        |
+-----------------------+------------+-----------+----------------+----------------------+
| TableReader_5         | 1500000.00 | root      |                | data:TableFullScan_4 |
| └─TableFullScan_4     | 1500000.00 | cop[tikv] | table:customer | keep order:false     |
+-----------------------+------------+-----------+----------------+----------------------+
2 rows in set (0.00 sec)
	`
	expected := [][2]int{{2, 24}, {26, 37}, {39, 49}, {51, 66}, {68, 89}}
	cols, _, _, err := GetHeaderColsPosition(planText, false, 0)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if !reflect.DeepEqual(cols, expected) {
		t.Fatalf("Expected %v, but got: %v", expected, cols)
	}
}

func TestGetHeaderColsPositionReturnsErrorWhenHeaderNotFound(t *testing.T) {
	planText := "No header here"
	_, _, _, err := GetHeaderColsPosition(planText, false, 0)
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}
}

func TestGetHeaderColsPositionReturnsErrorWhenPlanTextIsEmpty(t *testing.T) {
	planText := ""
	_, _, _, err := GetHeaderColsPosition(planText, false, 0)
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}
}
