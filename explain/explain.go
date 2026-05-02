package explain

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PlanNode stores one TiDB explain plan operator.
type PlanNode struct {
	ID            string      `json:"id"`           // 节点ID
	EstCost       float64     `json:"estCost"`      // 预估成本
	EstRows       float64     `json:"estRows"`      // 预估行数
	ActRows       float64     `json:"actRows"`      // 实际行数
	Task          string      `json:"taskType"`     // 任务名称，如：root, cop[tikv]等
	AccessObject  string      `json:"accessObject"` // 访问对象
	OperatorInfo  string      `json:"operatorInfo"` // 算子信息
	ExecutionInfo string      `json:"executeInfo"`  // 执行信息
	Memory        int         `json:"memoryInfo"`   // 内存信息，单位 byte
	Disk          int         `json:"diskInfo"`     // 磁盘信息，单位 byte
	MemoryBytes   int64       `json:"memoryBytes"`  // 内存信息，单位 byte
	DiskBytes     int64       `json:"diskBytes"`    // 磁盘信息，单位 byte
	PlanType      FormatType  `json:"planType"`     // 执行计划类型
	Children      []*PlanNode `json:"children"`     // 子节点，保留 TiDB explain 输出顺序
	planFlag      PlanFlag
	deep          int       // flag 距离行首的 rune 数
	Parent        *PlanNode `json:"-"` // 父节点
	Left          *PlanNode `json:"-"` // 兼容旧 API：第一个子节点
	Right         *PlanNode `json:"-"` // 兼容旧 API：第二个子节点
}

// getDeep returns the tree marker depth by rune index instead of byte index.
func (p *PlanNode) getDeep() int {
	if p.deep != 0 {
		return p.deep
	}
	if p.getPlanFlag() == RootFlag {
		return 0
	}
	re := regexp.MustCompile(`(└─|├─)`)
	idx := re.FindStringIndex(p.ID)
	if idx == nil {
		return 0
	}
	p.deep = len([]rune(p.ID[:idx[1]]))
	return p.deep
}

// IsBuildSide reports whether the operator is annotated as a join build side.
func (p *PlanNode) IsBuildSide() bool {
	return strings.Contains(p.ID, "Build")
}

func (p *PlanNode) getPlanFlag() PlanFlag {
	if p.planFlag != "" {
		return p.planFlag
	}
	if strings.Contains(p.ID, "├─") {
		p.planFlag = StartFlag
	} else if strings.Contains(p.ID, "└─") {
		p.planFlag = EndFlag
	} else {
		p.planFlag = RootFlag
	}
	return p.planFlag
}

func (p *PlanNode) Traverse() {
	fmt.Printf("PlanID:%s,Executor:%s,EstRows:%.2f\n", p.ID, p.GetExecutor(), p.EstRows)
	for _, child := range p.Children {
		child.Traverse()
	}
}

// IsLeaf reports whether the node has no children.
func (p *PlanNode) IsLeaf() bool {
	return len(p.Children) == 0
}

// AddChildren attaches a child to the receiver. It is kept for compatibility;
// NewPlanTree uses a stack-based builder for the full tree.
func (p *PlanNode) AddChildren(newChild *PlanNode) error {
	if p == nil || newChild == nil {
		return errors.New("nil plan node")
	}
	p.appendChild(newChild)
	return nil
}

func (p *PlanNode) appendChild(child *PlanNode) {
	child.Parent = p
	p.Children = append(p.Children, child)
	if len(p.Children) == 1 {
		p.Left = child
	} else if len(p.Children) == 2 {
		p.Right = child
	}
}

// GetExecutor returns the physical executor name from the id column.
func (p *PlanNode) GetExecutor() string {
	executor, err := getOperatorName(p.ID)
	if err != nil {
		return ""
	}
	return executor
}

type PlanRowParser func(row []string) (*PlanNode, error)

var formatParsers = map[FormatType]PlanRowParser{
	FormatTypePlanBriefText:   parsePlanBriefRow,
	FormatTypePlanVerboseText: parsePlanVerboseRow,
	FormatTypeAnalyzeVerboseText: func(row []string) (*PlanNode, error) {
		// explain analyze format='verbose' 的执行计划，在 select tidb_decode_binary_plan(BINARY_PLAN) from STATEMENTS_SUMMARY 中获取的也是这种格式
		if len(row) < 10 {
			return nil, fmt.Errorf("invalid analyze verbose row length: %d", len(row))
		}
		estRows, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, err
		}
		estCost, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, err
		}
		actRows, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, err
		}
		memInfo, err := parseUnit(row[8])
		if err != nil {
			return nil, err
		}
		diskInfo, err := parseUnit(row[9])
		if err != nil {
			return nil, err
		}
		return &PlanNode{
			ID:            row[0],
			EstRows:       estRows,
			EstCost:       estCost,
			ActRows:       actRows,
			Task:          row[4],
			AccessObject:  row[5],
			ExecutionInfo: row[6],
			OperatorInfo:  row[7],
			Memory:        int(memInfo),
			Disk:          int(diskInfo),
			MemoryBytes:   int64(memInfo),
			DiskBytes:     int64(diskInfo),
		}, nil
	},
}

// RegisterPlanRowParser allows callers to register custom row parsers for new formats.
func RegisterPlanRowParser(format FormatType, parser PlanRowParser) error {
	if parser == nil {
		return errors.New("parser is nil")
	}
	if _, exists := formatParsers[format]; exists {
		return fmt.Errorf("parser already registered for format %d", format)
	}
	formatParsers[format] = parser
	return nil
}

// NewPlanTree creates a tree from a raw TiDB explain plan.
func NewPlanTree(rawPlan *RawPlan) (planNode *PlanNode, err error) {
	if rawPlan == nil {
		return nil, errors.New("raw plan is nil")
	}
	if rawPlan.data == nil || len(rawPlan.data) == 0 {
		return nil, errors.New("raw plan is empty")
	}
	parser, ok := formatParsers[rawPlan.Tp]
	if !ok {
		return nil, errors.New("unsupported format type")
	}
	return buildPlanTree(rawPlan, parser)
}

func buildPlanTree(rawPlan *RawPlan, parser PlanRowParser) (*PlanNode, error) {
	var rootNode *PlanNode
	var stack []*PlanNode
	for i, row := range rawPlan.data {
		node, err := parser(row)
		if err != nil {
			return nil, err
		}
		node.PlanType = rawPlan.Tp
		depth := node.getDeep()
		if i == 0 {
			if node.getPlanFlag() != RootFlag {
				return nil, errors.New("first plan node is not root")
			}
			rootNode = node
			stack = []*PlanNode{node}
			continue
		}
		parentIdx := -1
		for j := len(stack) - 1; j >= 0; j-- {
			if stack[j].getDeep() < depth {
				parentIdx = j
				break
			}
		}
		if parentIdx < 0 {
			return nil, fmt.Errorf("cannot find parent for plan node %q", node.ID)
		}
		parent := stack[parentIdx]
		parent.appendChild(node)
		stack = append(stack[:parentIdx+1], node)
	}
	return rootNode, nil
}

func parsePlanBriefRow(row []string) (*PlanNode, error) {
	if len(row) < 5 {
		return nil, fmt.Errorf("invalid plan brief row length: %d", len(row))
	}
	estRows, err := strconv.ParseFloat(row[1], 64)
	if err != nil {
		return nil, err
	}
	return &PlanNode{
		ID:           row[0],
		EstRows:      estRows,
		Task:         row[2],
		AccessObject: row[3],
		OperatorInfo: row[4],
	}, nil
}

func parsePlanVerboseRow(row []string) (*PlanNode, error) {
	if len(row) < 6 {
		return nil, fmt.Errorf("invalid plan verbose row length: %d", len(row))
	}
	estRows, err := strconv.ParseFloat(row[1], 64)
	if err != nil {
		return nil, err
	}
	estCost, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return nil, err
	}
	return &PlanNode{
		ID:           row[0],
		EstRows:      estRows,
		EstCost:      estCost,
		Task:         row[3],
		AccessObject: row[4],
		OperatorInfo: row[5],
	}, nil
}
