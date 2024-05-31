package explain

import (
	"errors"
	"fmt"
	"log"
	"regexp"
)

type PlanFlag string

const StartFlag PlanFlag = "├─"  //占4个字节，当遇到该标志时，说明该行父节点还有兄弟节点
const EndFlag PlanFlag = "└─"    //占2个字节，当遇到该标志时，说明该行父节点已经是最后一个节点
const RootFlag PlanFlag = "root" //根节点标

type FormatType int

const (
	FormatTypePlanBriefText   FormatType = iota //执行计划简要文本，按行输出默认执行计划
	FormatTypePlanVerboseText                   //执行计划详细文本，包含成本预估信息，按行输出详细执行计划
	FormatTypePlanBriefJSON
	FormatTypePlanVerboseJSON
	FormatTypeAnalyzeBriefText   //执行计划分析简要文本，包含执行信息，按行输出默认执行计划
	FormatTypeAnalyzeVerboseText //执行计划分析详细文本，包含成本预估信息，执行信息，按行输出详细执行计划
	FormatTypeAnalyzeBriefJSON
	FormatTypeAnalyzeVerboseJSON
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
	deep          int        //flag距离行首的字节数
	childDeep     int        //子节点的deep，确定会存在右节点时，记录左节点的deep，临时使用
	planFlag      PlanFlag   //flag类型
	Parent        *PlanNode  //父节点
	Left          *PlanNode  //左子节点
	Right         *PlanNode  //右子节点
	line          string     //该行内容
	Executor      string     //算子名称
}

func (p *PlanNode) Traverse() {
	fmt.Println(p.GetExecutor())
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

	//前序遍历，遍历根节点，左子树，右子树
	if p.deep < newChild.deep {
		if newChild.planFlag == StartFlag {
			if p.IsLeaf() {
				p.childDeep = newChild.deep
				newChild.Parent = p
				p.Left = newChild
				log.Println("新增节点:", p.GetExecutor(), p.deep, p.childDeep, newChild.GetExecutor(), newChild.deep)
				return nil
			}
		} else if newChild.planFlag == EndFlag {
			//log.Println("判断右节点能否添加:", p.childDeep, newChild.deep, newChild.GetExecutor())
			if p.Left != nil && p.childDeep == newChild.deep {
				log.Println("newChild:", newChild.GetExecutor())
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
	if p.Executor != "" {
		return p.Executor
	}
	var re *regexp.Regexp
	// 提取Projection_28算子名称
	if p.planFlag == RootFlag {
		re = regexp.MustCompile(`\|\s*(?P<executor>\S+)\s+\|`)
	} else {
		re = regexp.MustCompile(`(└─|├─)(?P<executor>\S+)\s+\|`)
	}
	match := re.FindStringSubmatch(p.line)
	if len(match) == 0 {
		return ""
	}
	for i, name := range re.SubexpNames() {
		if name == "executor" {
			p.Executor = match[i]
			break
		}
	}
	return p.Executor
}
