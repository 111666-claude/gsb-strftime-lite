package strftimelite

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func mustFormat(t *testing.T, layout string, moment time.Time) string {
	t.Helper()
	got, err := Format(layout, moment)
	if err != nil {
		t.Fatalf("Format(%q) 报错：%v", layout, err)
	}
	return got
}

func TestZeroPadding(t *testing.T) {
	moment := time.Unix(1770000000, 0).UTC() // 2026-02-02 02:40:00 UTC
	if got := mustFormat(t, "%Y-%m-%d %H:%M:%S", moment); got != "2026-02-02 02:40:00" {
		t.Fatalf("补零不对：%q", got)
	}
}

func TestYearIsFourDigits(t *testing.T) {
	moment := time.Date(999, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := mustFormat(t, "%Y-%m-%d %H:%M:%S", moment); got != "0999-01-02 03:04:05" {
		t.Fatalf("年份应补零到四位：%q", got)
	}
}

func TestPercentEscape(t *testing.T) {
	moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	if got := mustFormat(t, "100%% done", moment); got != "100% done" {
		t.Fatalf("%%%% 应还原为一个百分号：%q", got)
	}
}

func TestUnknownDirective(t *testing.T) {
	moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	_, err := Format("ab%q", moment)
	if err == nil {
		t.Fatal("未知指令应报错")
	}
	if !strings.Contains(err.Error(), "%q") || !strings.Contains(err.Error(), "4") {
		t.Fatalf("错误应指出指令与位置（第 4 个字符）：%v", err)
	}
	if _, err := Parse("ab%q", "ab00"); err == nil {
		t.Fatal("解析方向未知指令也应报错")
	}
}

func TestYearOnly(t *testing.T) {
	moment := time.Date(2026, 2, 2, 2, 40, 0, 0, time.UTC)
	if got := mustFormat(t, "%Y", moment); got != "2026" {
		t.Fatalf("年份不对：%q", got)
	}
}

func TestPlainTextIsKept(t *testing.T) {
	moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	if got := mustFormat(t, "hello", moment); got != "hello" {
		t.Fatalf("普通文本应原样输出：%q", got)
	}
}

func TestTrailingPercent(t *testing.T) {
	moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	if got := mustFormat(t, "100%", moment); got != "100%" {
		t.Fatalf("末尾百分号应原样：%q", got)
	}
	parsed, err := Parse("100%", "100%")
	if err != nil || parsed != time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("末尾百分号解析应按字面匹配：%v %v", parsed, err)
	}
}

func TestLayoutList(t *testing.T) {
	joined := strings.Join(Layout(), ",")
	for _, directive := range []string{"%Y", "%m", "%d", "%H", "%M", "%S", "%%"} {
		if !strings.Contains(joined, directive) {
			t.Fatalf("指令表缺少 %s：%s", directive, joined)
		}
	}
}

func TestMonthAndDayAreNumbers(t *testing.T) {
	moment := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	got := mustFormat(t, "%m/%d", moment)
	if got != "12/25" {
		t.Fatalf("月日不对：%q", got)
	}
}

func TestParseBasic(t *testing.T) {
	moment, err := Parse("%Y-%m-%d %H:%M:%S", "2026-02-02 02:40:00")
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if moment.Unix() != 1770000000 {
		t.Fatalf("unix 秒不对：%d", moment.Unix())
	}
	if moment.Location() != time.UTC {
		t.Fatalf("应按 UTC 处理：%v", moment.Location())
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name   string
		layout string
		input  string
	}{
		{"字面不匹配", "%Y-%m", "2026/02"},
		{"多余字符", "%Y", "2026x"},
		{"非数字", "%m", "ab"},
		{"数字不足", "%Y", "202"},
		{"输入过短", "%Y-%m", "2026"},
		{"时越界", "%H", "24"},
		{"分越界", "%M", "60"},
		{"秒越界", "%S", "60"},
		{"月越界", "%m", "13"},
		{"日越界", "%d", "00"},
		{"非法日期", "%Y-%m-%d", "2026-02-30"},
	}
	for _, tc := range cases {
		if _, err := Parse(tc.layout, tc.input); err == nil {
			t.Fatalf("%s：应报错（layout=%q input=%q）", tc.name, tc.layout, tc.input)
		}
	}
}

func TestParseErrorPosition(t *testing.T) {
	_, err := Parse("%Y-%m-%d", "2026-02-x1")
	if err == nil || !strings.Contains(err.Error(), "9") {
		t.Fatalf("应指出第 9 个字符：%v", err)
	}
	_, err = Parse("%Y", "2026多余")
	if err == nil || !strings.Contains(err.Error(), "5") {
		t.Fatalf("多余内容应指出第 5 个字符：%v", err)
	}
}

func TestFormatIsDeterministic(t *testing.T) {
	moment := time.Unix(1770000000, 0).UTC()
	first := mustFormat(t, "%Y-%m-%d %H:%M:%S %%", moment)
	for i := 0; i < 1000; i++ {
		if got := mustFormat(t, "%Y-%m-%d %H:%M:%S %%", moment); got != first {
			t.Fatalf("重复格式化结果不一致：%q != %q", got, first)
		}
	}
}

// goLayout 把本引擎格式串翻译成 time.Format 的等价参考格式串。
func goLayout(layout string) string {
	var b strings.Builder
	for i := 0; i < len(layout); i++ {
		if layout[i] != '%' {
			b.WriteByte(layout[i])
			continue
		}
		if i+1 >= len(layout) {
			b.WriteByte('%')
			continue
		}
		switch layout[i+1] {
		case 'Y':
			b.WriteString("2006")
		case 'm':
			b.WriteString("01")
		case 'd':
			b.WriteString("02")
		case 'H':
			b.WriteString("15")
		case 'M':
			b.WriteString("04")
		case 'S':
			b.WriteString("05")
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(layout[i+1])
		}
		i++
	}
	return b.String()
}

// randomLayout 生成随机格式串。字面量刻意避开数字与 Go 参考格式串的魔法词。
func randomLayout(rng *rand.Rand) string {
	pieces := []string{
		"%Y", "%m", "%d", "%H", "%M", "%S", "%%",
		"-", "/", ":", " ", ".", ",", "T", "at", "done", "x",
	}
	n := 1 + rng.Intn(6)
	var b strings.Builder
	prev := ""
	for i := 0; i < n; i++ {
		piece := pieces[rng.Intn(len(pieces))]
		// Go 参考实现会把秒后的 "." / "," 当小数秒，避开以免对照失真。
		if prev == "%S" && (piece == "." || piece == ",") {
			piece = "-"
		}
		b.WriteString(piece)
		prev = piece
	}
	return b.String()
}

func TestAgainstReferenceImplementation(t *testing.T) {
	rng := rand.New(rand.NewSource(20260202))
	for i := 0; i < 10000; i++ {
		sec := int64(631152000) + rng.Int63n(631152000) // 1990~2010，秒级
		moment := time.Unix(sec, 0).UTC()
		layout := randomLayout(rng)
		got, err := Format(layout, moment)
		if err != nil {
			t.Fatalf("Format(%q) 报错：%v", layout, err)
		}
		want := moment.Format(goLayout(layout))
		if got != want {
			t.Fatalf("格式化不一致 layout=%q：%q != %q", layout, got, want)
		}
		parsed, err := Parse(layout, got)
		if err != nil {
			t.Fatalf("Parse(%q, %q) 报错：%v", layout, got, err)
		}
		ref, err := time.Parse(goLayout(layout), got)
		if err != nil {
			t.Fatalf("参考实现解析失败 layout=%q input=%q：%v", layout, got, err)
		}
		if parsed.Unix() != ref.Unix() {
			t.Fatalf("解析结果不一致 layout=%q input=%q：%d != %d", layout, got, parsed.Unix(), ref.Unix())
		}
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	layouts := []string{
		"%Y-%m-%d %H:%M:%S",
		"%Y/%m/%dT%H:%M:%S",
		"%Y%m%d%H%M%S",
		"%H:%M:%S %Y-%m-%d",
		"%Y-%m-%d",
	}
	for i := 0; i < 10000; i++ {
		sec := rng.Int63n(4102444800) // 1970~2100
		moment := time.Unix(sec, 0).UTC()
		layout := layouts[rng.Intn(len(layouts))]
		text := mustFormat(t, layout, moment)
		parsed, err := Parse(layout, text)
		if err != nil {
			t.Fatalf("往返失败 layout=%q text=%q：%v", layout, text, err)
		}
		if !strings.Contains(layout, "%S") {
			continue // 缺秒级字段时无法保证秒级不变量
		}
		if parsed.Unix() != moment.Unix() {
			t.Fatalf("不变量不成立 layout=%q：%d != %d", layout, parsed.Unix(), moment.Unix())
		}
	}
}

func TestFormatScalesLinearly(t *testing.T) {
	moment := time.Unix(1770000000, 0).UTC()
	const n = 100000
	for i := 0; i < n; i++ {
		got, err := Format("%Y-%m-%d %H:%M:%S", moment)
		if err != nil || got != "2026-02-02 02:40:00" {
			t.Fatalf("第 %d 次格式化不对：%q %v", i, got, err)
		}
	}
}
