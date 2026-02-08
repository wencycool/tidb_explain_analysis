package explain

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/wencycool/tidb_explain_analysis/explain/plancodec"
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

var (
	operatorNameRegexp = regexp.MustCompile(`(└─|├─){0,1}(?P<executor>\w+)(_\d+){1}\s*`)
	unitMultipliers    = map[string]float64{
		"B":     1,
		"Bytes": 1,
		"KB":    1024,
		"MB":    1024 * 1024,
		"GB":    1024 * 1024 * 1024,
		"TB":    1024 * 1024 * 1024 * 1024,
	}
)

// 判断一个字符串中是否包含算子名称，如果存在则返回算子名称，否则返回空字符串
func getOperatorName(line string) (string, error) {
	match := operatorNameRegexp.FindStringSubmatch(line)
	var executor string
	if len(match) == 0 {
		executor = ""
	} else {
		for i, name := range operatorNameRegexp.SubexpNames() {
			if name == "executor" {
				executor = match[i]
				break
			}
		}
	}
	if plancodec.TypeStringToPhysicalID(executor) == 0 {
		return "", errors.New("invalid executor name")
	}
	return executor, nil
}

// 解析memory和disk信息，将其变为字节
func parseUnit(str string) (float64, error) {
	/*
	   " 12.1 KB"," 1.07 MB"," N/A"," 0 Bytes","12.3 GB"等多种形式
	*/
	str = strings.TrimSpace(str)
	if str == "" || str == "N/A" {
		return 0, nil
	}
	parts := strings.Fields(str)
	if len(parts) == 1 {
		value, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	if len(parts) != 2 {
		return 0, errors.New("invalid unit")
	}
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	multiplier, ok := unitMultipliers[parts[1]]
	if !ok {
		return 0, errors.New("invalid unit")
	}
	return value * multiplier, nil
}
