package tidb_explain_analysis

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Format identifies a TiDB explain output format supported by this package.
type Format string

const (
	FormatText          Format = "text"
	FormatVerboseText   Format = "verbose_text"
	FormatAnalyzeText   Format = "analyze_text"
	FormatAnalyzeVerbose Format = "analyze_verbose_text"
)

// Plan is a parsed TiDB execution plan plus all flattened nodes.
type Plan struct {
	Format Format      `json:"format"`
	Root   *Node       `json:"root"`
	Nodes  []*Node     `json:"nodes"`
	Rows   []PlanRow   `json:"rows"`
}

// Node is one physical operator in a TiDB execution plan.
type Node struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Depth         int           `json:"depth"`
	EstRows       float64       `json:"estRows,omitempty"`
	EstCost       float64       `json:"estCost,omitempty"`
	ActRows       float64       `json:"actRows,omitempty"`
	Task          string        `json:"task,omitempty"`
	AccessObject  string        `json:"accessObject,omitempty"`
	OperatorInfo  string        `json:"operatorInfo,omitempty"`
	ExecutionInfo string        `json:"executionInfo,omitempty"`
	MemoryBytes   int64         `json:"memoryBytes,omitempty"`
	DiskBytes     int64         `json:"diskBytes,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
	Loops         int64         `json:"loops,omitempty"`
	Parent        *Node         `json:"-"`
	Children      []*Node       `json:"children,omitempty"`
}

func (n *Node) Walk(fn func(*Node)) {
	if n == nil || fn == nil {
		return
	}
	fn(n)
	for _, child := range n.Children {
		child.Walk(fn)
	}
}

func (n *Node) RowEstimationRatio() float64 {
	if n == nil || n.EstRows <= 0 || n.ActRows <= 0 {
		return 0
	}
	if n.ActRows > n.EstRows {
		return n.ActRows / n.EstRows
	}
	return n.EstRows / n.ActRows
}

func (p *Plan) Findings(rules ...Rule) []Finding {
	return Analyze(p, rules...)
}

func (p *Plan) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// PlanRow is one row decoded from TiDB's text table output.
type PlanRow map[string]string

// ParseOption configures parsing behavior.
type ParseOption func(*parseConfig)

type parseConfig struct {
	format Format
}

func WithFormat(format Format) ParseOption {
	return func(c *parseConfig) { c.format = format }
}

// ParseText parses TiDB EXPLAIN / EXPLAIN ANALYZE text output into a tree.
func ParseText(text string, opts ...ParseOption) (*Plan, error) {
	cfg := parseConfig{format: FormatAnalyzeVerbose}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	rows, err := parseTextRows(text)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no plan rows found")
	}
	plan := &Plan{Format: cfg.format, Rows: rows}
	for _, row := range rows {
		node, err := rowToNode(row)
		if err != nil {
			return nil, err
		}
		plan.Nodes = append(plan.Nodes, node)
	}
	if err := buildTree(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func buildTree(plan *Plan) error {
	var stack []*Node
	for i, node := range plan.Nodes {
		if i == 0 {
			plan.Root = node
			stack = []*Node{node}
			continue
		}
		parentIdx := -1
		for j := len(stack) - 1; j >= 0; j-- {
			if stack[j].Depth < node.Depth {
				parentIdx = j
				break
			}
		}
		if parentIdx < 0 {
			return fmt.Errorf("cannot find parent for node %q", node.ID)
		}
		parent := stack[parentIdx]
		node.Parent = parent
		parent.Children = append(parent.Children, node)
		stack = append(stack[:parentIdx+1], node)
	}
	return nil
}

func rowToNode(row PlanRow) (*Node, error) {
	id := rowValue(row, "id")
	name, depth := parseNodeID(id)
	if name == "" {
		return nil, fmt.Errorf("invalid plan node id %q", id)
	}
	n := &Node{
		ID:            strings.TrimSpace(id),
		Name:          name,
		Depth:         depth,
		Task:          rowValue(row, "task"),
		AccessObject:  rowValue(row, "access object"),
		OperatorInfo:  rowValue(row, "operator info"),
		ExecutionInfo: rowValue(row, "execution info"),
	}
	n.EstRows = parseFloat(rowValue(row, "estRows"))
	n.EstCost = parseFloat(rowValue(row, "estCost"))
	n.ActRows = parseFloat(rowValue(row, "actRows"))
	n.MemoryBytes = parseBytes(rowValue(row, "memory"))
	n.DiskBytes = parseBytes(rowValue(row, "disk"))
	n.Duration = parseExecutionDuration(n.ExecutionInfo)
	n.Loops = parseExecutionLoops(n.ExecutionInfo)
	return n, nil
}

func rowValue(row PlanRow, key string) string {
	return strings.TrimSpace(row[strings.ToLower(key)])
}

var nodeNameRegexp = regexp.MustCompile(`(?m)(?:└─|├─)?(?P<name>[A-Za-z][A-Za-z0-9]*)(?:_\d+)`)

func parseNodeID(id string) (string, int) {
	match := nodeNameRegexp.FindStringSubmatch(id)
	if len(match) == 0 {
		return "", 0
	}
	nameIdx := nodeNameRegexp.SubexpIndex("name")
	markerIdx := strings.Index(id, "├─")
	if markerIdx < 0 {
		markerIdx = strings.Index(id, "└─")
	}
	depth := 0
	if markerIdx >= 0 {
		depth = len([]rune(id[:markerIdx])) + 1
	}
	return match[nameIdx], depth
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func parseBytes(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "N/A") {
		return 0
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0
	}
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	if len(parts) == 1 {
		return int64(value)
	}
	units := map[string]float64{"B": 1, "Bytes": 1, "KB": 1024, "MB": 1024 * 1024, "GB": 1024 * 1024 * 1024, "TB": 1024 * 1024 * 1024 * 1024}
	return int64(value * units[parts[1]])
}

var durationRegexp = regexp.MustCompile(`time:([0-9.]+)(ns|µs|us|ms|s)`) 

func parseExecutionDuration(s string) time.Duration {
	m := durationRegexp.FindStringSubmatch(s)
	if len(m) != 3 {
		return 0
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	switch m[2] {
	case "ns":
		return time.Duration(v)
	case "µs", "us":
		return time.Duration(v * float64(time.Microsecond))
	case "ms":
		return time.Duration(v * float64(time.Millisecond))
	case "s":
		return time.Duration(v * float64(time.Second))
	default:
		return 0
	}
}

var loopsRegexp = regexp.MustCompile(`loops:([0-9]+)`) 

func parseExecutionLoops(s string) int64 {
	m := loopsRegexp.FindStringSubmatch(s)
	if len(m) != 2 {
		return 0
	}
	v, _ := strconv.ParseInt(m[1], 10, 64)
	return v
}

func parseTextRows(text string) ([]PlanRow, error) {
	lines := strings.Split(text, "\n")
	headerIdx := -1
	var headers []string
	for i, line := range lines {
		cols := splitTableLine(line)
		if hasPlanHeader(cols) {
			headerIdx = i
			headers = normalizeHeaders(cols)
			break
		}
	}
	if headerIdx < 0 {
		return nil, fmt.Errorf("plan header not found")
	}
	var rows []PlanRow
	for _, line := range lines[headerIdx+1:] {
		cols := splitTableLine(line)
		if len(cols) < len(headers) || len(cols) == 0 {
			continue
		}
		if _, _ = parseNodeID(cols[0]); parseNodeIDName(cols[0]) == "" {
			continue
		}
		row := PlanRow{}
		for i, h := range headers {
			row[h] = strings.TrimSpace(cols[i])
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseNodeIDName(id string) string {
	name, _ := parseNodeID(id)
	return name
}

func splitTableLine(line string) []string {
	line = strings.TrimRight(line, "\r")
	if !strings.Contains(line, "|") {
		return nil
	}
	first := strings.Index(line, "|")
	last := strings.LastIndex(line, "|")
	if first == last {
		return nil
	}
	parts := strings.Split(line[first+1:last], "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func hasPlanHeader(cols []string) bool {
	seen := map[string]bool{}
	for _, c := range normalizeHeaders(cols) {
		seen[c] = true
	}
	return seen["id"] && seen["estRows"] && seen["task"] && seen["operator info"]
}

func normalizeHeaders(cols []string) []string {
	out := make([]string, 0, len(cols))
	for _, c := range cols {
		out = append(out, strings.ToLower(strings.TrimSpace(c)))
	}
	return out
}

func SortFindings(findings []Finding) {
	severityRank := map[Severity]int{SeverityCritical: 0, SeverityWarning: 1, SeverityInfo: 2}
	sort.SliceStable(findings, func(i, j int) bool {
		if severityRank[findings[i].Severity] != severityRank[findings[j].Severity] {
			return severityRank[findings[i].Severity] < severityRank[findings[j].Severity]
		}
		return findings[i].NodeID < findings[j].NodeID
	})
}
