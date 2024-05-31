package explain

import "errors"

type plan [][]string // 保留执行计划的原始数据，用于text格式的分析

type RawPlan struct {
	Tp FormatType // 执行计划类型
	P  plan       // 执行计划的原始数据
}

// 从文本中获取执行计划

func (r *RawPlan) GetPlanFromText(text string, formatType FormatType) (*RawPlan, error) {
	var p plan
	switch formatType {
	case FormatTypePlanBriefText, FormatTypePlanVerboseText, FormatTypeAnalyzeBriefText, FormatTypeAnalyzeVerboseText:
		//p = getPlanFromText(text)
	default:
		return nil, errors.New("unsupported format type")
	}
	return &RawPlan{Tp: formatType, P: p}, nil

}

// FormatTypePlanBriefText 格式的执行计划
