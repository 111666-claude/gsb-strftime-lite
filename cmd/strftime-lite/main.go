// Command strftime-lite 按格式串打印时间。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"example.com/strftimelite"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("strftime-lite", flag.ContinueOnError)
	flags.SetOutput(stderr)
	layout := flags.String("layout", "%Y-%m-%d %H:%M:%S", "格式串")
	at := flags.Int64("at", 1770000000, "unix 秒（UTC）")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	moment := time.Unix(*at, 0).UTC()
	fmt.Fprintln(stdout, strftimelite.Format(*layout, moment))
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
