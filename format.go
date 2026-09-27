// Package strftimelite 做时间格式串引擎（格式化与解析），只用标准库。
package strftimelite

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Format 按格式串输出 UTC 时间。数字字段补零（%Y 四位，其余两位），
// %% 输出一个百分号，末尾孤立的 % 原样输出，未知指令报错并指出位置。
func Format(layout string, moment time.Time) (string, error) {
	moment = moment.UTC()
	var out strings.Builder
	out.Grow(len(layout) + 8)
	for index := 0; index < len(layout); index++ {
		if layout[index] != '%' {
			out.WriteByte(layout[index])
			continue
		}
		if index+1 >= len(layout) {
			out.WriteByte('%')
			continue
		}
		switch layout[index+1] {
		case 'Y':
			writePadded(&out, moment.Year(), 4)
		case 'm':
			writePadded(&out, int(moment.Month()), 2)
		case 'd':
			writePadded(&out, moment.Day(), 2)
		case 'H':
			writePadded(&out, moment.Hour(), 2)
		case 'M':
			writePadded(&out, moment.Minute(), 2)
		case 'S':
			writePadded(&out, moment.Second(), 2)
		case '%':
			out.WriteByte('%')
		default:
			return "", fmt.Errorf("未知指令 %%%c（格式串第 %d 个字符）", layout[index+1], index+2)
		}
		index++
	}
	return out.String(), nil
}

// Parse 按格式串解析输入，返回 UTC 时间（秒级精度）。
// 不匹配、多余字符、非法数字、越界字段都会报错并给出位置。
func Parse(layout, input string) (time.Time, error) {
	year, month, day := 0, 1, 1
	hour, minute, second := 0, 0, 0
	pos := 0
	for index := 0; index < len(layout); index++ {
		if layout[index] != '%' {
			if err := expectByte(input, &pos, layout[index]); err != nil {
				return time.Time{}, err
			}
			continue
		}
		if index+1 >= len(layout) {
			if err := expectByte(input, &pos, '%'); err != nil {
				return time.Time{}, err
			}
			continue
		}
		directive := layout[index+1]
		index++
		if directive == '%' {
			if err := expectByte(input, &pos, '%'); err != nil {
				return time.Time{}, err
			}
			continue
		}
		var width int
		var target *int
		switch directive {
		case 'Y':
			width, target = 4, &year
		case 'm':
			width, target = 2, &month
		case 'd':
			width, target = 2, &day
		case 'H':
			width, target = 2, &hour
		case 'M':
			width, target = 2, &minute
		case 'S':
			width, target = 2, &second
		default:
			return time.Time{}, fmt.Errorf("未知指令 %%%c（格式串第 %d 个字符）", directive, index+1)
		}
		value, err := readDigits(input, &pos, width)
		if err != nil {
			return time.Time{}, err
		}
		*target = value
	}
	if pos != len(input) {
		return time.Time{}, fmt.Errorf("输入第 %d 个字符起有多余内容", pos+1)
	}
	if err := checkRange("月", month, 1, 12); err != nil {
		return time.Time{}, err
	}
	if err := checkRange("日", day, 1, 31); err != nil {
		return time.Time{}, err
	}
	if err := checkRange("时", hour, 0, 23); err != nil {
		return time.Time{}, err
	}
	if err := checkRange("分", minute, 0, 59); err != nil {
		return time.Time{}, err
	}
	if err := checkRange("秒", second, 0, 59); err != nil {
		return time.Time{}, err
	}
	moment := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	if moment.Year() != year || int(moment.Month()) != month || moment.Day() != day {
		return time.Time{}, fmt.Errorf("非法日期：%04d-%02d-%02d 不存在", year, month, day)
	}
	return moment, nil
}

// writePadded 十进制输出 value，不足 width 位左侧补 0。
func writePadded(out *strings.Builder, value, width int) {
	digits := strconv.Itoa(value)
	for i := len(digits); i < width; i++ {
		out.WriteByte('0')
	}
	out.WriteString(digits)
}

// expectByte 要求输入当前位置恰好是 want，否则报错并指出位置。
func expectByte(input string, pos *int, want byte) error {
	if *pos >= len(input) {
		return fmt.Errorf("输入第 %d 个字符处期望 %q，但输入已结束", *pos+1, want)
	}
	if input[*pos] != want {
		return fmt.Errorf("输入第 %d 个字符不匹配：期望 %q，得到 %q", *pos+1, want, input[*pos])
	}
	*pos++
	return nil
}

// readDigits 从输入读 width 位十进制数字。
func readDigits(input string, pos *int, width int) (int, error) {
	if *pos+width > len(input) {
		return 0, fmt.Errorf("输入第 %d 个字符处需要 %d 位数字，但输入已结束", *pos+1, width)
	}
	value := 0
	for k := 0; k < width; k++ {
		c := input[*pos+k]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("输入第 %d 个字符不是数字：%q", *pos+k+1, c)
		}
		value = value*10 + int(c-'0')
	}
	*pos += width
	return value, nil
}

// checkRange 校验字段取值范围。
func checkRange(name string, value, low, high int) error {
	if value < low || value > high {
		return fmt.Errorf("%s越界：%d 不在 [%d, %d]", name, value, low, high)
	}
	return nil
}

// Layout 返回支持的指令表。
func Layout() []string {
	return []string{"%Y", "%m", "%d", "%H", "%M", "%S", "%%"}
}
