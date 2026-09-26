// Package numfmt 按十进制规则格式化数值。
package numfmt

import (
	"fmt"
	"strings"
)

// Format 把十进制字符串 value 格式化为保留 digits 位小数的字符串。
// 舍入按输入字符串逐位进行（四舍六入五成双），不经过二进制浮点；
// group 为 true 时整数部分每三位加逗号；结果为零时不带负号。
func Format(value string, digits int, group bool) (string, error) {
	if digits < 0 || digits > 10 {
		return "", fmt.Errorf("digits 超出范围 [0,10]：%d（输入 %q）", digits, value)
	}
	negative, intPart, fracPart, ok := splitDecimal(strings.TrimSpace(value))
	if !ok {
		return "", fmt.Errorf("不是合法的数值：%q", value)
	}

	intDigits, fracDigits := roundHalfEven(intPart, fracPart, digits)

	if isZero(intDigits) && isZero(fracDigits) {
		negative = false
	}
	intDigits = strings.TrimLeft(intDigits, "0")
	if intDigits == "" {
		intDigits = "0"
	}
	if group {
		intDigits = groupThousands(intDigits)
	}

	var builder strings.Builder
	if negative {
		builder.WriteByte('-')
	}
	builder.WriteString(intDigits)
	if digits > 0 {
		builder.WriteByte('.')
		builder.WriteString(fracDigits)
	}
	return builder.String(), nil
}

// splitDecimal 把十进制字符串拆成符号、整数部分、小数部分，并校验只含数字。
func splitDecimal(text string) (negative bool, intPart, fracPart string, ok bool) {
	if text == "" {
		return false, "", "", false
	}
	if text[0] == '+' || text[0] == '-' {
		negative = text[0] == '-'
		text = text[1:]
	}
	intPart = text
	if dot := strings.IndexByte(text, '.'); dot >= 0 {
		intPart, fracPart = text[:dot], text[dot+1:]
		if strings.IndexByte(fracPart, '.') >= 0 {
			return false, "", "", false
		}
	}
	if intPart == "" && fracPart == "" {
		return false, "", "", false
	}
	if !allDigits(intPart) || !allDigits(fracPart) {
		return false, "", "", false
	}
	if intPart == "" {
		intPart = "0"
	}
	return negative, intPart, fracPart, true
}

// roundHalfEven 对 intPart.fracPart 按四舍六入五成双保留 digits 位小数，
// 返回整数部分与恰好 digits 位的小数部分（整数部分可能因进位多一位）。
func roundHalfEven(intPart, fracPart string, digits int) (string, string) {
	keep := len(intPart) + digits
	all := make([]byte, 0, keep+1)
	all = append(all, intPart...)
	all = append(all, fracPart...)

	switch {
	case len(all) < keep:
		for len(all) < keep {
			all = append(all, '0')
		}
	case len(all) > keep:
		dropped := all[keep:]
		all = all[:keep]
		if roundUp(dropped, all[keep-1]) {
			all = increment(all)
		}
	}
	return string(all[:len(all)-digits]), string(all[len(all)-digits:])
}

// roundUp 按四舍六入五成双判断被舍弃的首位数字是否触发进位。
// dropped 非空且只含数字；lastKept 是保留部分的最后一位。
func roundUp(dropped []byte, lastKept byte) bool {
	switch {
	case dropped[0] > '5':
		return true
	case dropped[0] < '5':
		return false
	}
	for _, digit := range dropped[1:] {
		if digit != '0' {
			return true
		}
	}
	// 恰好舍弃一半：五成双，保留位为奇数才进位。
	return lastKept%2 == 1
}

// increment 把十进制数字串加一，溢出时在前面补 1。
func increment(digits []byte) []byte {
	for i := len(digits) - 1; i >= 0; i-- {
		if digits[i] < '9' {
			digits[i]++
			return digits
		}
		digits[i] = '0'
	}
	return append([]byte{'1'}, digits...)
}

// groupThousands 给整数字符串每三位加逗号。
func groupThousands(intDigits string) string {
	n := len(intDigits)
	if n <= 3 {
		return intDigits
	}
	commas := (n - 1) / 3
	first := n - commas*3
	buf := make([]byte, 0, n+commas)
	buf = append(buf, intDigits[:first]...)
	for i := first; i < n; i += 3 {
		buf = append(buf, ',')
		buf = append(buf, intDigits[i:i+3]...)
	}
	return string(buf)
}

func allDigits(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func isZero(digits string) bool {
	return strings.TrimLeft(digits, "0") == ""
}
