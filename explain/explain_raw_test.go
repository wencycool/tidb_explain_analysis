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
	expected := [][2]int{{2, 25}, {26, 38}, {39, 50}, {51, 67}, {68, 90}}
	cols, _, err := GetHeaderColsPosition(planText, false)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if !reflect.DeepEqual(cols, expected) {
		t.Fatalf("Expected %v, but got: %v", expected, cols)
	}
}

func TestGetHeaderColsPositionReturnsErrorWhenHeaderNotFound(t *testing.T) {
	planText := "No header here"
	_, _, err := GetHeaderColsPosition(planText, false)
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}
}

func TestGetHeaderColsPositionReturnsErrorWhenPlanTextIsEmpty(t *testing.T) {
	planText := ""
	_, _, err := GetHeaderColsPosition(planText, false)
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}
}
