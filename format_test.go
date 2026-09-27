package strftimelite

import (
"strings"
"testing"
"time"
)

func TestYearOnly(t *testing.T) {
moment := time.Date(2026, 2, 2, 2, 40, 0, 0, time.UTC)
if got := Format("%Y", moment); got != "2026" {
t.Fatalf("年份不对：%q", got)
}
}

func TestPlainTextIsKept(t *testing.T) {
moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
if got := Format("hello", moment); got != "hello" {
t.Fatalf("普通文本应原样输出：%q", got)
}
}

func TestTrailingPercent(t *testing.T) {
moment := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
if got := Format("100%", moment); got != "100%" {
t.Fatalf("末尾百分号应原样：%q", got)
}
}

func TestLayoutList(t *testing.T) {
joined := strings.Join(Layout(), ",")
for _, directive := range []string{"%Y", "%m", "%d", "%H", "%M", "%S"} {
if !strings.Contains(joined, directive) {
t.Fatalf("指令表缺少 %s：%s", directive, joined)
}
}
}

func TestMonthAndDayAreNumbers(t *testing.T) {
moment := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
got := Format("%m/%d", moment)
if !strings.Contains(got, "12") || !strings.Contains(got, "25") {
t.Fatalf("月日不对：%q", got)
}
}
