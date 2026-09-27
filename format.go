// Package strftimelite 做时间格式串引擎（格式化与解析），只用标准库。
package strftimelite

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Format 按格式串输出 UTC 时间。数字字段补零，%% 还原为一个百分号，
// 末尾孤立的 % 原样输出，未知指令报错并指出位置（1 起）。
func Format(layout string, moment time.Time) (string, error) {
	moment = moment.UTC()
	var out strings.Builder
	out.Grow(len(layout) + 8)
	for index := 0; index < len(layout); index++ {
		ch := layout[index]
		if ch != '%' {
			out.WriteByte(ch)
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
			return "", fmt.Errorf("未知指令 %%%c（第 %d 个字符）", layout[index+1], index+1)
		}
		index++
	}
	return out.String(), nil
}

// writePadded 十进制输出 value，不足 width 位左侧补 0。
func writePadded(out *strings.Builder, value, width int) {
	digits := strconv.Itoa(value)
	for i := len(digits); i < width; i++ {
		out.WriteByte('0')
	}
	out.WriteString(digits)
}

// Parse 按格式串解析 input，返回 UTC unix 秒。
// 不匹配、多余字符、非法数字、越界字段都报错并给出位置（1 起）。
// 未出现的字段取 time.Parse 同款默认值：月 1、日 1、其余 0。
func Parse(input, layout string) (int64, error) {
	year, month, day := 0, 1, 1
	hour, minute, second := 0, 0, 0
	pos := 0
	for index := 0; index < len(layout); index++ {
		ch := layout[index]
		if ch != '%' {
			if err := expectByte(input, &pos, ch); err != nil {
				return 0, err
			}
			continue
		}
		if index+1 >= len(layout) {
			if err := expectByte(input, &pos, '%'); err != nil {
				return 0, err
			}
			continue
		}
		directive := layout[index+1]
		if directive == '%' {
			if err := expectByte(input, &pos, '%'); err != nil {
				return 0, err
			}
			index++
			continue
		}
		width := 0
		switch directive {
		case 'Y':
			width = 4
		case 'm', 'd', 'H', 'M', 'S':
			width = 2
		default:
			return 0, fmt.Errorf("未知指令 %%%c（第 %d 个字符）", directive, index+1)
		}
		index++
		fieldPos := pos
		value, err := readDigits(input, &pos, width)
		if err != nil {
			return 0, err
		}
		switch directive {
		case 'Y':
			year = value
		case 'm':
			if value < 1 || value > 12 {
				return 0, fmt.Errorf("第 %d 个字符：月份 %02d 越界（1-12）", fieldPos+1, value)
			}
			month = value
		case 'd':
			if value < 1 || value > 31 {
				return 0, fmt.Errorf("第 %d 个字符：日 %02d 越界（1-31）", fieldPos+1, value)
			}
			day = value
		case 'H':
			if value > 23 {
				return 0, fmt.Errorf("第 %d 个字符：时 %02d 越界（0-23）", fieldPos+1, value)
			}
			hour = value
		case 'M':
			if value > 59 {
				return 0, fmt.Errorf("第 %d 个字符：分 %02d 越界（0-59）", fieldPos+1, value)
			}
			minute = value
		case 'S':
			if value > 59 {
				return 0, fmt.Errorf("第 %d 个字符：秒 %02d 越界（0-59）", fieldPos+1, value)
			}
			second = value
		}
	}
	if pos != len(input) {
		return 0, fmt.Errorf("第 %d 个字符起有多余内容：%q", pos+1, input[pos:])
	}
	moment := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	if moment.Year() != year || int(moment.Month()) != month || moment.Day() != day {
		return 0, fmt.Errorf("非法日期：%04d-%02d-%02d 不存在", year, month, day)
	}
	return moment.Unix(), nil
}

// expectByte 要求 input 在 *pos 处正好是 want。
func expectByte(input string, pos *int, want byte) error {
	if *pos >= len(input) {
		return fmt.Errorf("第 %d 个字符处期望 %q，但输入已结束", *pos+1, want)
	}
	if input[*pos] != want {
		return fmt.Errorf("第 %d 个字符不匹配：期望 %q，得到 %q", *pos+1, want, input[*pos])
	}
	*pos++
	return nil
}

// readDigits 从 *pos 起读 width 位十进制数字。
func readDigits(input string, pos *int, width int) (int, error) {
	if *pos+width > len(input) {
		return 0, fmt.Errorf("第 %d 个字符处需要 %d 位数字，但输入已结束", *pos+1, width)
	}
	value := 0
	for k := 0; k < width; k++ {
		digit := input[*pos+k]
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("第 %d 个字符不是数字：%q", *pos+k+1, digit)
		}
		value = value*10 + int(digit-'0')
	}
	*pos += width
	return value, nil
}

// Layout 返回支持的指令表。
func Layout() []string {
	return []string{"%Y", "%m", "%d", "%H", "%M", "%S", "%%"}
}
