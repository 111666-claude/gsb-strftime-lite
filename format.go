// Package strftimelite 做时间格式串引擎。
package strftimelite

import (
	"fmt"
	"time"
)

// Format 按格式串输出。缺陷：数字字段不补零，%% 也不当转义。
func Format(layout string, moment time.Time) string {
	out := ""
	for index := 0; index < len(layout); index++ {
		if layout[index] != '%' || index+1 >= len(layout) {
			out += string(layout[index])
			continue
		}
		switch layout[index+1] {
		case 'Y':
			out += fmt.Sprintf("%d", moment.Year())
		case 'm':
			out += fmt.Sprintf("%d", int(moment.Month()))
		case 'd':
			out += fmt.Sprintf("%d", moment.Day())
		case 'H':
			out += fmt.Sprintf("%d", moment.Hour())
		case 'M':
			out += fmt.Sprintf("%d", moment.Minute())
		case 'S':
			out += fmt.Sprintf("%d", moment.Second())
		case '%':
			// 缺陷：%% 没有还原成一个百分号
			out += "%%"
		default:
			out += string(layout[index+1])
		}
		index++
	}
	return out
}

// Layout 返回支持的指令表。
func Layout() []string {
	return []string{"%Y", "%m", "%d", "%H", "%M", "%S", "%%"}
}
