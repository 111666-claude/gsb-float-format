// Package numfmt 按十进制规则格式化数值。
package numfmt

import (
	"fmt"
	"strings"
)

// maxDigits 是支持的最大小数位数。
const maxDigits = 10

// Format 按十进制字符串逐位舍入（四舍六入五成双），保留 digits 位小数；
// group 为真时整数部分每三位加逗号；结果为零时不带负号。
func Format(value string, digits int, group bool) (string, error) {
	if digits < 0 || digits > maxDigits {
		return "", fmt.Errorf("不支持的位数 %d（仅支持 0 到 %d）：%q", digits, maxDigits, value)
	}
	trimmed := strings.TrimSpace(value)
	body := trimmed
	negative := false
	if body != "" && (body[0] == '+' || body[0] == '-') {
		negative = body[0] == '-'
		body = body[1:]
	}
	intPart, fracPart, ok := splitDecimal(body)
	if !ok {
		return "", fmt.Errorf("不是合法的数值：%q", value)
	}
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}
	intPart, fracPart = roundHalfEven(intPart, fracPart, digits)
	if isZero(intPart) && isZero(fracPart) {
		negative = false
	}
	if group {
		intPart = groupThousands(intPart)
	}
	var out strings.Builder
	out.Grow(len(intPart) + len(fracPart) + 2)
	if negative {
		out.WriteByte('-')
	}
	out.WriteString(intPart)
	if digits > 0 {
		out.WriteByte('.')
		out.WriteString(fracPart)
	}
	return out.String(), nil
}

// splitDecimal 把无符号十进制串拆成整数与小数部分，要求至少有一位数字。
func splitDecimal(s string) (intPart, fracPart string, ok bool) {
	if s == "" {
		return "", "", false
	}
	digits := 0
	dot := -1
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c >= '0' && c <= '9':
			digits++
		case c == '.' && dot < 0:
			dot = i
		default:
			return "", "", false
		}
	}
	if digits == 0 {
		return "", "", false
	}
	if dot < 0 {
		return s, "", true
	}
	return s[:dot], s[dot+1:], true
}

// roundHalfEven 把小数部分截到 digits 位，按四舍六入五成双进位，必要时进位到整数部分。
func roundHalfEven(intPart, fracPart string, digits int) (string, string) {
	if len(fracPart) > digits {
		kept := fracPart[:digits]
		dropped := fracPart[digits:]
		lastKept := byte('0')
		if digits > 0 {
			lastKept = kept[digits-1]
		} else {
			lastKept = intPart[len(intPart)-1]
		}
		if roundUp(dropped, lastKept) {
			if digits == 0 {
				var carry bool
				intPart, carry = increment(intPart)
				if carry {
					intPart = "1" + intPart
				}
			} else {
				var carry bool
				kept, carry = increment(kept)
				if carry {
					intPart, carry = increment(intPart)
					if carry {
						intPart = "1" + intPart
					}
				}
			}
		}
		fracPart = kept
	}
	if len(fracPart) < digits {
		fracPart += strings.Repeat("0", digits-len(fracPart))
	}
	return intPart, fracPart
}

// roundUp 依据被丢弃的首位、其后是否有非零位、保留末位的奇偶决定是否进位。
func roundUp(dropped string, lastKept byte) bool {
	first := dropped[0]
	if first != '5' {
		return first > '5'
	}
	for i := 1; i < len(dropped); i++ {
		if dropped[i] != '0' {
			return true
		}
	}
	return lastKept%2 == 1
}

// increment 把十进制数字串加一，返回新串与是否溢出进位；空串视为 0。
func increment(s string) (string, bool) {
	buf := []byte(s)
	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] != '9' {
			buf[i]++
			return string(buf), false
		}
		buf[i] = '0'
	}
	if len(buf) == 0 {
		return "1", false
	}
	return string(buf), true
}

// isZero 报告十进制串是否全部为零。
func isZero(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}

// groupThousands 给整数字符串每三位加逗号。
func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var out strings.Builder
	out.Grow(n + (n-1)/3)
	lead := n % 3
	if lead > 0 {
		out.WriteString(s[:lead])
	}
	for i := lead; i < n; i += 3 {
		if i > 0 {
			out.WriteByte(',')
		}
		out.WriteString(s[i : i+3])
	}
	return out.String()
}
