// Command float-format 按十进制规则输出数值。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/numfmt"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("float-format", flag.ContinueOnError)
	flags.SetOutput(stderr)
	value := flags.String("format", "", "要格式化的十进制字符串")
	digits := flags.Int("digits", 2, "保留几位小数")
	group := flags.Bool("group", false, "整数部分加千分位")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *value == "" {
		fmt.Fprintln(stderr, "用法：float-format --format 2.675 --digits 2 [--group]")
		return 2
	}
	text, err := numfmt.Format(*value, *digits, *group)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, text)
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
