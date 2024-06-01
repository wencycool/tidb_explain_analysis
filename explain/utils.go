package explain

import (
	"errors"
	"gitee.com/wencycool/tidb_explain_analyze/plancodec"
	"regexp"
	"strconv"
	"strings"
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

// 判断一个字符串中是否包含算子名称，如果存在则返回算子名称，否则返回空字符串
func getOperatorName(line string) (string, error) {
	var re *regexp.Regexp
	re = regexp.MustCompile(`(└─|├─){0,1}(?P<executor>\w+)(_\d+){1}\s*`)
	match := re.FindStringSubmatch(line)
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
	var unit float64
	str = strings.TrimSpace(str)
	if str == "N/A" {
		return 0, nil
	}
	if strings.HasSuffix(str, "Bytes") {
		dStr := strings.TrimSpace(strings.TrimSuffix(str, "Bytes"))
		unit, err := strconv.ParseFloat(dStr, 64)
		if err != nil {
			return 0, err
		} else {
			return unit, nil
		}
	} else if strings.HasSuffix(str, "KB") {
		dStr := strings.TrimSpace(strings.TrimSuffix(str, "KB"))
		d, err := strconv.ParseFloat(dStr, 64)
		if err != nil {
			return 0, err
		} else {
			unit = d * 1024
		}
		return unit, nil
	} else if strings.HasSuffix(str, "MB") {
		dStr := strings.TrimSpace(strings.TrimSuffix(str, "MB"))
		d, err := strconv.ParseFloat(dStr, 64)
		if err != nil {
			return 0, err
		} else {
			unit = d * 1024 * 1024
		}
		return unit, nil
	} else if strings.HasSuffix(str, "GB") {
		dStr := strings.TrimSpace(strings.TrimSuffix(str, "GB"))
		d, err := strconv.ParseFloat(dStr, 64)
		if err != nil {
			return 0, err
		} else {
			unit = d * 1024 * 1024 * 1024
		}
		return unit, nil
	} else {
		return 0, errors.New("invalid unit")
	}
}
