package strftimelite

import (
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
	moment := time.Unix(1770000000, 0).UTC() // 2026-02-02 02:40:00
	if got := mustFormat(t, "%Y-%m-%d %H:%M:%S", moment); got != "2026-02-02 02:40:00" {
		t.Fatalf("补零不对：%q", got)
	}
}

func TestYearOnly(t *testing.T) {
	moment := time.Date(2026, 2, 2, 2, 40, 0, 0, time.UTC)
	if got := mustFormat(t, "%Y", moment); got != "2026" {
		t.Fatalf("年份不对：%q", got)
	}
}

func TestYearZeroPaddedToFourDigits(t *testing.T) {
	moment := time.Date(26, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := mustFormat(t, "%Y", moment); got != "0026" {
		t.Fatalf("年份应补零到四位：%q", got)
	}
}

func TestPercentEscape(t *testing.T) {
	moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	if got := mustFormat(t, "100%% done", moment); got != "100% done" {
		t.Fatalf("%%%% 应还原为一个百分号：%q", got)
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
}

func TestUnknownDirectiveReportsPosition(t *testing.T) {
	moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	_, err := Format("ab%q", moment)
	if err == nil || !strings.Contains(err.Error(), "第 3 个字符") {
		t.Fatalf("未知指令应报错并指出位置：%v", err)
	}
	if _, err := Parse("ab", "ab%q"); err == nil || !strings.Contains(err.Error(), "第 3 个字符") {
		t.Fatalf("解析方向未知指令也应报错：%v", err)
	}
}

func TestParseRoundTrip(t *testing.T) {
	moment := time.Unix(1770000000, 0).UTC()
	layout := "%Y-%m-%d %H:%M:%S"
	text := mustFormat(t, layout, moment)
	seconds, err := Parse(text, layout)
	if err != nil {
		t.Fatalf("Parse(%q) 报错：%v", text, err)
	}
	if seconds != 1770000000 {
		t.Fatalf("往返不变量不成立：%d != 1770000000", seconds)
	}
}

func TestParseErrors(t *testing.T) {
	layout := "%Y-%m-%d %H:%M:%S"
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"字面量不匹配", "2026/02-02 02:40:00", "第 5 个字符"},
		{"多余字符", "2026-02-02 02:40:00xyz", "第 20 个字符"},
		{"非数字", "2026-a2-02 02:40:00", "第 6 个字符"},
		{"输入过短", "2026-02-0", "输入已结束"},
		{"秒越界", "2026-02-02 02:40:60", "越界（0-59）"},
		{"分越界", "2026-02-02 02:60:00", "越界（0-59）"},
		{"时越界", "2026-02-02 24:40:00", "越界（0-23）"},
		{"月越界", "2026-13-02 02:40:00", "越界（1-12）"},
		{"日越界", "2026-02-32 02:40:00", "越界（1-31）"},
		{"非法日期", "2026-02-30 02:40:00", "非法日期"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.input, layout)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse(%q) 错误应包含 %q：%v", tc.input, tc.want, err)
			}
		})
	}
}

func TestParseTrailingPercentLayout(t *testing.T) {
	seconds, err := Parse("100%", "100%")
	if err != nil {
		t.Fatalf("末尾孤立 %% 应作为字面量解析：%v", err)
	}
	if seconds != time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Unix() {
		t.Fatalf("默认值不对：%d", seconds)
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

func TestFormatScalesLinearly(t *testing.T) {
	moment := time.Unix(1770000000, 0).UTC()
	layout := "%Y-%m-%d %H:%M:%S"
	want := mustFormat(t, layout, moment)
	for i := 0; i < 100000; i++ {
		if got := mustFormat(t, layout, moment); got != want {
			t.Fatalf("第 %d 次格式化结果不一致", i)
		}
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
