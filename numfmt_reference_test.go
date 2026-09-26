package numfmt

import (
	"math/rand"
	"strings"
	"testing"
)

// bankerTable[丢弃首位][丢弃位后是否有非零][保留末位为奇] = 是否进位。
// 四舍六入五成双的完整判定表，供参考实现查表使用。
var bankerTable = [10][2][2]bool{
	{},                            // 0
	{},                            // 1
	{},                            // 2
	{},                            // 3
	{},                            // 4
	{{false, true}, {true, true}}, // 5：无尾数看奇偶，有尾数必进
	{{true, true}, {true, true}},  // 6
	{{true, true}, {true, true}},  // 7
	{{true, true}, {true, true}},  // 8
	{{true, true}, {true, true}},  // 9
}

// referenceFormat 是独立的手工十进制舍入参考实现：把输入拆成符号与数字序列，
// 查银行家舍入表决定是否进位，再手工做加一与千分位拼接。
func referenceFormat(value string, digits int, group bool) (string, error) {
	s := strings.TrimSpace(value)
	neg := false
	if s != "" && (s[0] == '+' || s[0] == '-') {
		neg = s[0] == '-'
		s = s[1:]
	}
	dot := strings.IndexByte(s, '.')
	var intStr, fracStr string
	if dot < 0 {
		intStr = s
	} else {
		intStr, fracStr = s[:dot], s[dot+1:]
		if strings.IndexByte(fracStr, '.') >= 0 {
			return "", errInvalid(value)
		}
	}
	if !allDigits(intStr) || !allDigits(fracStr) || len(intStr)+len(fracStr) == 0 {
		return "", errInvalid(value)
	}
	intStr = strings.TrimLeft(intStr, "0")
	if intStr == "" {
		intStr = "0"
	}
	// 进位判定：查表。
	if len(fracStr) > digits {
		dropped := fracStr[digits:]
		tail := false
		for i := 1; i < len(dropped); i++ {
			if dropped[i] != '0' {
				tail = true
				break
			}
		}
		lastKept := byte('0')
		if digits > 0 {
			lastKept = fracStr[digits-1]
		} else {
			lastKept = intStr[len(intStr)-1]
		}
		odd := 0
		if lastKept%2 == 1 {
			odd = 1
		}
		hasTail := 0
		if tail {
			hasTail = 1
		}
		if bankerTable[dropped[0]-'0'][hasTail][odd] {
			intStr, fracStr = refIncrement(intStr, fracStr[:digits])
		} else {
			fracStr = fracStr[:digits]
		}
	}
	for len(fracStr) < digits {
		fracStr += "0"
	}
	if neg && allZero(intStr) && allZero(fracStr) {
		neg = false
	}
	var out strings.Builder
	if neg {
		out.WriteByte('-')
	}
	if group {
		for i, c := range intStr {
			if i > 0 && (len(intStr)-i)%3 == 0 {
				out.WriteByte(',')
			}
			out.WriteRune(c)
		}
	} else {
		out.WriteString(intStr)
	}
	if digits > 0 {
		out.WriteByte('.')
		out.WriteString(fracStr)
	}
	return out.String(), nil
}

// refIncrement 对「整数部分.小数部分」整体加一（小数位对齐到末位），返回新串。
func refIncrement(intStr, fracStr string) (string, string) {
	digits := []byte(intStr + fracStr)
	i := len(digits) - 1
	for ; i >= 0; i-- {
		if digits[i] != '9' {
			digits[i]++
			break
		}
		digits[i] = '0'
	}
	if i < 0 {
		digits = append([]byte{'1'}, digits...)
		intStr = string(digits[:len(digits)-len(fracStr)])
	} else {
		intStr = string(digits[:len(intStr)])
	}
	return intStr, string(digits[len(digits)-len(fracStr):])
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func allZero(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}

type invalidError string

func (e invalidError) Error() string { return "不是合法的数值：" + string(e) }

func errInvalid(value string) error { return invalidError(strings.TrimSpace(value)) }

func TestAgainstReferenceSamples(t *testing.T) {
	samples := []string{
		"2.675", "2.5", "3.5", "0.125", "0.5", "1.5", "-0.001", "-0",
		"1234567.891", "-1234567.891", "9.999", "0.995", "999.999",
		"0.045", "2.685", "2.695", "10.5", "11.5", "1", "1.2", "  7  ",
		"0.10000000005", "99999.999995",
	}
	for _, in := range samples {
		for digits := 0; digits <= 10; digits++ {
			for _, group := range []bool{false, true} {
				got, err := Format(in, digits, group)
				want, refErr := referenceFormat(in, digits, group)
				if (err != nil) != (refErr != nil) {
					t.Fatalf("Format(%q,%d,%v) 错误状态不一致：%v vs %v", in, digits, group, err, refErr)
				}
				if err == nil && got != want {
					t.Fatalf("Format(%q,%d,%v) = %q；参考实现给 %q", in, digits, group, got, want)
				}
			}
		}
	}
}

func TestAgainstReferenceRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(20260927))
	for i := 0; i < 10000; i++ {
		var sb strings.Builder
		if rng.Intn(2) == 0 {
			sb.WriteByte('-')
		}
		intLen := rng.Intn(12)
		for j := 0; j < intLen; j++ {
			sb.WriteByte(byte('0' + rng.Intn(10)))
		}
		fracLen := rng.Intn(13)
		if fracLen > 0 || rng.Intn(2) == 0 {
			sb.WriteByte('.')
			for j := 0; j < fracLen; j++ {
				sb.WriteByte(byte('0' + rng.Intn(10)))
			}
		}
		in := sb.String()
		digits := rng.Intn(11)
		group := rng.Intn(2) == 0
		got, err := Format(in, digits, group)
		want, refErr := referenceFormat(in, digits, group)
		if (err != nil) != (refErr != nil) {
			t.Fatalf("第 %d 例 Format(%q,%d,%v) 错误状态不一致：%v vs %v", i, in, digits, group, err, refErr)
		}
		if err == nil && got != want {
			t.Fatalf("第 %d 例 Format(%q,%d,%v) = %q；参考实现给 %q", i, in, digits, group, got, want)
		}
	}
}
