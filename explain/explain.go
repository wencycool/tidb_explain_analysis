package explain

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// tidb目前只支持二叉树的join，所以这里只需要考虑二叉树的情况
// 存放explain <sql>的执行计划

type PlanNode struct {
	ID            string     `json:"id"`           //节点ID
	EstCost       float64    `json:"estCost"`      //预估成本
	EstRows       float64    `json:"estRows"`      //预估行数
	ActRows       float64    `json:"actRows"`      //实际行数
	Task          string     `json:"taskType"`     //任务名称，如：root, cop[tikv]等
	AccessObject  string     `json:"accessObject"` //访问对象
	OperatorInfo  string     `json:"operatorInfo"` //算子信息
	ExecutionInfo string     `json:"executeInfo"`  //执行信息
	Memory        int        `json:"memoryInfo"`   //内存信息
	Disk          int        `json:"diskInfo"`     //磁盘信息
	PlanType      FormatType //执行计划类型
	planFlag      PlanFlag
	deep          int       //flag距离行首的字节数
	childDeep     int       //子节点的deep，确定会存在右节点时，记录左节点的deep，临时使用
	Parent        *PlanNode //父节点
	Left          *PlanNode //左子节点
	Right         *PlanNode //右子节点
}

// 获取flag距离行首的字节数
func (p *PlanNode) getDeep() int {
	if p.deep != 0 {
		return p.deep
	}
	// 计算flag距离行首的字节数
	if p.getPlanFlag() == RootFlag {
		return 0
	}
	re := regexp.MustCompile(`(└─|├─)`)
	pos := re.FindStringIndex(p.ID)
	if pos != nil {
		p.deep = pos[1]
		return p.deep
	}
	log.Println("get deep failed:", p.ID)
	return 0
}

// 判断当前节点时build端还是probe端

func (p *PlanNode) IsBuildSide() bool {
	if strings.Contains(p.ID, "Build") {
		return true
	}
	return false
}

func (p *PlanNode) getPlanFlag() PlanFlag {
	if p.planFlag != "" {
		return p.planFlag
	}
	// 通过解析ID字段，判断当前节点的flag
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
	if p.Left != nil {
		p.Left.Traverse()
	}
	if p.Right != nil {
		p.Right.Traverse()
	}
}

// 判断当前节点是否是叶子节点

func (p *PlanNode) IsLeaf() bool {
	return p.Left == nil && p.Right == nil
}

func (p *PlanNode) AddChildren(newChild *PlanNode) error {
	if p.getPlanFlag() == RootFlag && p.Left == nil {
		newChild.Parent = p
		p.Left = newChild
		return nil
	}
	//前序遍历，遍历根节点，左子树，右子树
	if p.getDeep() < newChild.getDeep() {
		if newChild.getPlanFlag() == StartFlag {
			if p.IsLeaf() {
				p.childDeep = newChild.getDeep()
				newChild.Parent = p
				p.Left = newChild
				return nil
			}
		} else if newChild.getPlanFlag() == EndFlag {
			//log.Println("判断右节点能否添加:", p.childDeep, newChild.deep, newChild.GetExecutor())
			if p.Left != nil && p.childDeep == newChild.getDeep() {
				newChild.Parent = p
				p.Right = newChild
				p.childDeep = 0
				return nil
			} else if p.IsLeaf() {
				newChild.Parent = p
				p.Left = newChild
				return nil
			}
		}
	}
	if p.Right != nil {
		return p.Right.AddChildren(newChild)
	} else if p.Left != nil {
		return p.Left.AddChildren(newChild)
	}
	return errors.New("add children failed")
}

// " |       └─Projection_28                              |"
// 利用正则表达式找到算子名称

func (p *PlanNode) GetExecutor() string {
	var re *regexp.Regexp
	// 提取Projection_28算子名称
	if p.getPlanFlag() == RootFlag {
		re = regexp.MustCompile(`\s*(?P<executor>\w+)(_\d+){1}\s*`)
	} else {
		re = regexp.MustCompile(`(└─|├─)(?P<executor>\w+)(_\d+){1}\s*`)
	}
	match := re.FindStringSubmatch(p.ID)
	var executor string
	if len(match) == 0 {
		executor = ""
	} else {
		for i, name := range re.SubexpNames() {
			if name == "executor" {
				executor = match[i]
				break
			}
		}
	}
	return executor
}

// 创建执行计划树
// todo 目前只支持FormatTypePlanBriefText

func NewPlanTree(rawPlan *RawPlan) (planNode *PlanNode, err error) {
	if rawPlan == nil {
		return nil, errors.New("raw plan is nil")
	}
	if rawPlan.data == nil {
		return nil, errors.New("raw plan is empty")
	}
	if len(rawPlan.data) == 0 {
		return nil, errors.New("raw plan is empty")
	}
	var rootNode *PlanNode
	if rawPlan.Tp != FormatTypePlanBriefText {
		return nil, errors.New("unsupported format type")
	}
	for i, row := range rawPlan.data {
		estRows, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, err
		}
		tmpNode := &PlanNode{
			ID:           row[0],
			EstRows:      estRows,
			Task:         row[2],
			AccessObject: row[3],
			OperatorInfo: row[4],
			PlanType:     rawPlan.Tp,
		}
		if i == 0 {
			rootNode = tmpNode
		} else {
			if err = rootNode.AddChildren(tmpNode); err != nil {
				return nil, err
			}
		}
	}
	return rootNode, nil
}
