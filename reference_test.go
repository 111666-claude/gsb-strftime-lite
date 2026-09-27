package strftimelite

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

// goReference 是我们的指令到 time.Format 等价格式串的映射（参考实现）。
var goReference = map[byte]string{
	'Y': "2006",
	'm': "01",
	'd': "02",
	'H': "15",
	'M': "04",
	'S': "05",
}

// referenceFormat 参考实现：逐指令调用 time.Format，字面量原样拼接。
// 这样可以避开 Go 布局把数字字面量（如 "100" 里的 "1"）当成指令的坑。
func referenceFormat(t *testing.T, layout string, moment time.Time) string {
	t.Helper()
	var out strings.Builder
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
		directive := layout[index+1]
		if directive == '%' {
			out.WriteByte('%')
		} else if ref, ok := goReference[directive]; ok {
			out.WriteString(moment.Format(ref))
		} else {
			t.Fatalf("参考实现遇到未知指令 %%%c", directive)
		}
		index++
	}
	return out.String()
}

// referenceLayout 把格式串整体翻译成 Go 布局，供 time.Parse 对照用。
func referenceLayout(t *testing.T, layout string) string {
	t.Helper()
	var out strings.Builder
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
		directive := layout[index+1]
		if directive == '%' {
			out.WriteByte('%')
		} else if ref, ok := goReference[directive]; ok {
			out.WriteString(ref)
		} else {
			t.Fatalf("参考实现遇到未知指令 %%%c", directive)
		}
		index++
	}
	return out.String()
}

// goLayoutSafe 判断字面量是否会被 Go 布局误当成指令（含数字或月/周/AM-PM 名称）。
func goLayoutSafe(layout string) bool {
	tokens := []string{
		"0", "1", "2", "3", "4", "5", "6",
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
		"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun",
		"PM", "AM", "MST",
	}
	for index := 0; index < len(layout); index++ {
		ch := layout[index]
		if ch != '%' {
			for _, token := range tokens {
				if strings.HasPrefix(layout[index:], token) {
					return false
				}
			}
			continue
		}
		if index+1 < len(layout) {
			index++
		}
	}
	return true
}

func hasAllFields(layout string) bool {
	for _, directive := range []string{"%Y", "%m", "%d", "%H", "%M", "%S"} {
		if !strings.Contains(layout, directive) {
			return false
		}
	}
	return true
}

// TestAgainstReference 对样例与 1 万个随机「时间 + 格式串」逐字符比对
// Format 输出，并比对 Parse 结果（参考实现为 time.Format / time.Parse）。
func TestAgainstReference(t *testing.T) {
	layouts := []string{
		"%Y-%m-%d %H:%M:%S",
		"100%% done",
		"%Y/%m/%d",
		"%H:%M:%S",
		"at %H:%M on %Y-%m-%d (100%%)",
		"trailing %",
		"%S",
	}
	rng := rand.New(rand.NewSource(42))
	directives := []string{"%Y", "%m", "%d", "%H", "%M", "%S", "%%"}
	literals := []string{"-", ":", " ", "/", "T", "at ", "x", ""}
	for i := 0; i < 10000; i++ {
		var layout strings.Builder
		for k := 0; k < 1+rng.Intn(6); k++ {
			layout.WriteString(directives[rng.Intn(len(directives))])
			layout.WriteString(literals[rng.Intn(len(literals))])
		}
		layouts = append(layouts, layout.String())
	}
	for _, layout := range layouts {
		safe := goLayoutSafe(layout)
		goLayout := ""
		if safe {
			goLayout = referenceLayout(t, layout)
		}
		for i := 0; i < 20; i++ {
			seconds := rng.Int63n(253402300799) // 0000-9999 年，秒级
			moment := time.Unix(seconds, 0).UTC()

			got, err := Format(layout, moment)
			if err != nil {
				t.Fatalf("Format(%q) 报错：%v", layout, err)
			}
			want := referenceFormat(t, layout, moment)
			if got != want {
				t.Fatalf("Format(%q, %d) = %q，参考实现 = %q", layout, seconds, got, want)
			}

			gotSeconds, err := Parse(got, layout)
			if err != nil {
				t.Fatalf("Parse(%q, %q) 报错：%v", got, layout, err)
			}
			if safe {
				refMoment, err := time.Parse(goLayout, want)
				if err != nil {
					t.Fatalf("参考实现 time.Parse(%q, %q) 报错：%v", goLayout, want, err)
				}
				if gotSeconds != refMoment.Unix() {
					t.Fatalf("Parse(%q, %q) = %d，参考实现 = %d", got, layout, gotSeconds, refMoment.Unix())
				}
			} else if hasAllFields(layout) && gotSeconds != seconds {
				t.Fatalf("Parse(%q, %q) = %d，期望 %d", got, layout, gotSeconds, seconds)
			}
		}
	}
}

// TestRoundTripInvariant 对含全部字段的格式串验证 Parse(Format(t)) == t（秒级）。
func TestRoundTripInvariant(t *testing.T) {
	layout := "%Y-%m-%d %H:%M:%S"
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 10000; i++ {
		seconds := rng.Int63n(253402300799)
		moment := time.Unix(seconds, 123456789).UTC() // 纳秒应被秒级取整
		text, err := Format(layout, moment)
		if err != nil {
			t.Fatalf("Format 报错：%v", err)
		}
		got, err := Parse(text, layout)
		if err != nil {
			t.Fatalf("Parse(%q) 报错：%v", text, err)
		}
		if got != seconds {
			t.Fatalf("往返不变量不成立：%d -> %q -> %d", seconds, text, got)
		}
	}
}
