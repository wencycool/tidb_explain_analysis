package explain

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
