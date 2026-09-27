// Command strftime-lite 按格式串打印或解析时间。
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
	parse := flags.String("parse", "", "按格式串解析该时间串，输出 unix 秒")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *parse != "" {
		moment, err := strftimelite.Parse(*layout, *parse)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, moment.Unix())
		return 0
	}
	moment := time.Unix(*at, 0).UTC()
	out, err := strftimelite.Format(*layout, moment)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, out)
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
