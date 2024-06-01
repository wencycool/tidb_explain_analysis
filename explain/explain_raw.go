package explain

import (
	"errors"
	"strings"
)

type RawPlan struct {
	Tp   FormatType // 执行计划类型
	data [][]string // 执行计划的原始数据
}

// 从文本中获取执行计划

func GetRawPlanFromText(text string, formatType FormatType) (rawPlan *RawPlan, err error) {
	var data [][]string
	switch formatType {
	case FormatTypePlanBriefText:
		data, err = getPlanFromText(text)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported format type")
	}
	return &RawPlan{Tp: formatType, data: data}, nil

}

// FormatTypePlanBriefText 格式的执行计划

func getPlanFromText(planText string) (data [][]string, err error) {
	withShift := true
	colsPosition, shift, lineNo, err := GetHeaderColsPosition(planText, withShift, 0)
	if err != nil {
		return nil, err
	}
	flagCnt := 0
	for i, line := range strings.Split(planText, "\n") {
		// 跳过表头
		if i <= lineNo {
			continue
		}
		if strings.Contains(line, "+----") {
			flagCnt++
			continue
		}
		if flagCnt == 2 {
			return data, nil
		}
		var row []string
		// 按照rune进行截取，解析字符串
		lineRunes := []rune(line)
		var firstCol = true
		for j := 0; j < len(colsPosition); j++ {
			start := colsPosition[j][0] + shift
			end := colsPosition[j][1] + shift
			if end > len(lineRunes) {
				end = len(lineRunes)
			}
			var col string
			col = string(lineRunes[start:end])
			if firstCol {
				firstCol = false
			} else {
				col = strings.TrimSpace(string(lineRunes[start:end]))
			}
			row = append(row, col)
		}

		//// todo 是否会遇到字段中包含"|"的情况?
		//var firstCol = true
		//for _, c := range strings.Split(strings.TrimSpace(line), "|") {
		//	if c != "" {
		//		// 首列不能去空格，因为需要根据id的长度来判断层级
		//		if firstCol {
		//			firstCol = false
		//		} else {
		//			c = strings.TrimSpace(c)
		//		}
		//		row = append(row, c)
		//	}
		//}
		if len(data) > 0 {
			if len(row) != len(data[0]) {
				return nil, errors.New("row length not match")
			}
		}
		data = append(data, row)
	}
	return data, nil
}

/*
+-----------------------+------------+-----------+----------------+----------------------+
| id                    | estRows    | task      | access object  | operator info        |
+-----------------------+------------+-----------+----------------+----------------------+
| TableReader_5         | 1500000.00 | root      |                | data:TableFullScan_4 |
| └─TableFullScan_4     | 1500000.00 | cop[tikv] | table:customer | keep order:false     |
+-----------------------+------------+-----------+----------------+----------------------+
2 rows in set (0.00 sec)
*/

// 从文本中获表头的起始位置，用于做内容的字符串截取
// withShift 表示是否需要做偏移量的调整，如果是则表示每个字段的起始位置是相对于第一个"|"的偏移量
// 返回值cols是一个二维数组，每个元素是一个长度为2的数组，表示每个字段的起始位置，所有位置都是针对第一个"|"的偏移量
// lineNo是表头所在的行号
// shift是第一个"|"的偏移量
// 如果没有找到表头，返回错误

func GetHeaderColsPosition(planText string, withShift bool, splitFlag rune) (cols [][2]int, shift, lineNo int, err error) {
	// 首行必须全为ascii字符
	if splitFlag == 0 {
		splitFlag = '|'
	}
	for i, line := range strings.Split(planText, "\n") {
		// 判断是否全部包含多个字符串
		if strings.Contains(line, "id") && strings.Contains(line, "estRows") && strings.Contains(line, "task") && strings.Contains(line, "access object") && strings.Contains(line, "operator info") {
			if withShift {
				shift = strings.Index(line, string(splitFlag))
			}
			// 确认是header行，利用“|”分割字符串，获取每个字段的起始位置
			for _, c := range strings.Split(line[shift:len(line)], string(splitFlag)) {
				// 去掉\t和\r
				if c != "" && c != string(rune(9)) && c != string(rune(13)) {
					cols = append(cols, [2]int{strings.Index(line, c) - shift, strings.Index(line, c) + len(c) - shift})
				}
			}
			lineNo = i
			return cols, shift, lineNo, nil
		}
	}
	return nil, 0, 0, errors.New("header not found")
}
