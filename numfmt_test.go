package numfmt

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// roundOnFive 是银行家舍入表：恰好舍弃一个 5（后面全是 0）时，
// 保留的最后一位是奇数才进位（五成双）。
var roundOnFive = [10]bool{false, true, false, true, false, true, false, true, false, true}

// referenceFormat 是参照实现：按十进制字符串手工舍入，与 Format 的结构刻意不同，
// 用来逐字符对照生产实现的输出。
func referenceFormat(value string, digits int, group bool) (string, error) {
	if digits < 0 || digits > 10 {
		return "", fmt.Errorf("digits 超出范围 [0,10]：%d（输入 %q）", digits, value)
	}
	text := strings.TrimSpace(value)
	negative := false
	if strings.HasPrefix(text, "+") || strings.HasPrefix(text, "-") {
		negative = text[0] == '-'
		text = text[1:]
	}
	parts := strings.Split(text, ".")
	if len(parts) > 2 {
		return "", fmt.Errorf("不是合法的数值：%q", value)
	}
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}
	if intPart == "" && fracPart == "" {
		return "", fmt.Errorf("不是合法的数值：%q", value)
	}
	for _, ch := range intPart + fracPart {
		if ch < '0' || ch > '9' {
			return "", fmt.Errorf("不是合法的数值：%q", value)
		}
	}

	// 把整数与小数合成一个数字队列，保留位之后的首位决定舍入。
	kept := []byte(intPart)
	for len(kept) < digits {
		kept = append([]byte{'0'}, kept...)
	}
	frac := []byte(fracPart)
	carry := false
	if len(frac) > digits {
		first := frac[digits] - '0'
		rest := frac[digits+1:]
		switch {
		case first > 5:
			carry = true
		case first == 5:
			nonzero := false
			for _, digit := range rest {
				if digit != '0' {
					nonzero = true
					break
				}
			}
			if nonzero {
				carry = true
			} else {
				last := 0
				if digits > 0 {
					last = int(frac[digits-1] - '0')
				} else if len(intPart) > 0 {
					last = int(intPart[len(intPart)-1] - '0')
				}
				carry = roundOnFive[last]
			}
		}
		frac = frac[:digits]
	}
	for len(frac) < digits {
		frac = append(frac, '0')
	}
	kept = append(kept, frac...)
	if carry {
		// 手工进位：从右往左逐位加一。
		i := len(kept) - 1
		for ; i >= 0; i-- {
			if kept[i] != '9' {
				kept[i]++
				break
			}
			kept[i] = '0'
		}
		if i < 0 {
			kept = append([]byte{'1'}, kept...)
		}
	}
	intDigits := string(kept[:len(kept)-digits])
	fracDigits := string(kept[len(kept)-digits:])

	// 负零归一：整数与小数部分全是零则去掉负号。
	zero := true
	for _, ch := range intDigits + fracDigits {
		if ch != '0' {
			zero = false
			break
		}
	}
	if zero {
		negative = false
	}

	intDigits = strings.TrimLeft(intDigits, "0")
	if intDigits == "" {
		intDigits = "0"
	}
	if group {
		var grouped []byte
		for i, ch := range intDigits {
			if i > 0 && (len(intDigits)-i)%3 == 0 {
				grouped = append(grouped, ',')
			}
			grouped = append(grouped, byte(ch))
		}
		intDigits = string(grouped)
	}

	var out strings.Builder
	if negative {
		out.WriteByte('-')
	}
	out.WriteString(intDigits)
	if digits > 0 {
		out.WriteByte('.')
		out.WriteString(fracDigits)
	}
	return out.String(), nil
}

func checkFormat(t *testing.T, value string, digits int, group bool, want string) {
	t.Helper()
	got, err := Format(value, digits, group)
	if err != nil {
		t.Fatalf("Format(%q, %d, %v) 报错：%v", value, digits, group, err)
	}
	if got != want {
		t.Fatalf("Format(%q, %d, %v) = %q，期望 %q", value, digits, group, got, want)
	}
	again, err := Format(value, digits, group)
	if err != nil || again != got {
		t.Fatalf("Format(%q, %d, %v) 重复调用不一致：%q vs %q", value, digits, group, got, again)
	}
}

func TestFormatIntegerWithDigits(t *testing.T) {
	checkFormat(t, "1", 2, false, "1.00")
}

func TestFormatRoundsHalfToEven(t *testing.T) {
	checkFormat(t, "2.5", 0, false, "2")
	checkFormat(t, "3.5", 0, false, "4")
	checkFormat(t, "0.125", 2, false, "0.12")
	checkFormat(t, "0.135", 2, false, "0.14")
	checkFormat(t, "4.5", 0, false, "4")
	checkFormat(t, "5.5", 0, false, "6")
	checkFormat(t, "-2.5", 0, false, "-2")
	checkFormat(t, "-3.5", 0, false, "-4")
	checkFormat(t, "1.005", 2, false, "1.00")
	checkFormat(t, "1.015", 2, false, "1.02")
	checkFormat(t, "0.5001", 0, false, "1")
	checkFormat(t, "0.4999", 0, false, "0")
}

func TestFormatDecimalStringRounding(t *testing.T) {
	// 二进制浮点里 2.675 是 2.67499…，按十进制字符串舍入必须是 2.68。
	checkFormat(t, "2.675", 2, false, "2.68")
	checkFormat(t, "9.999", 2, false, "10.00")
	checkFormat(t, "99.5", 0, false, "100")
	checkFormat(t, "999.5", 0, false, "1000")
}

func TestFormatKeepsTrailingZeros(t *testing.T) {
	checkFormat(t, "1.2", 3, false, "1.200")
	checkFormat(t, "7", 0, false, "7")
	checkFormat(t, "0", 5, false, "0.00000")
	checkFormat(t, "1.23456789", 10, false, "1.2345678900")
}

func TestFormatGrouping(t *testing.T) {
	checkFormat(t, "1234567.891", 2, true, "1,234,567.89")
	checkFormat(t, "-1234567.891", 2, true, "-1,234,567.89")
	checkFormat(t, "123456789", 0, true, "123,456,789")
	checkFormat(t, "123", 2, true, "123.00")
	checkFormat(t, "1234", 0, true, "1,234")
	checkFormat(t, "0.5", 1, true, "0.5")
	checkFormat(t, "1234567.891", 2, false, "1234567.89")
}

func TestFormatNegativeZero(t *testing.T) {
	checkFormat(t, "-0.001", 2, false, "0.00")
	checkFormat(t, "-0", 0, false, "0")
	checkFormat(t, "-0.0", 3, false, "0.000")
	checkFormat(t, "-0.005", 2, false, "0.00")
	checkFormat(t, "-0.006", 2, false, "-0.01")
	checkFormat(t, "-0.0001", 2, true, "0.00")
}

func TestFormatRejectsGarbage(t *testing.T) {
	for _, bad := range []string{"abc", "", "1.2.3", "1e5", "--1", "1,000", ".", "+", "  ", "12a", "NaN", "Inf"} {
		if _, err := Format(bad, 2, false); err == nil {
			t.Fatalf("Format(%q) 应该报错", bad)
		} else if !strings.Contains(err.Error(), bad) && strings.TrimSpace(bad) != "" {
			t.Fatalf("Format(%q) 的错误信息应带原串：%v", bad, err)
		}
	}
}

func TestFormatDigitsRange(t *testing.T) {
	for _, digits := range []int{-1, 11, 100} {
		if _, err := Format("1.5", digits, false); err == nil {
			t.Fatalf("digits=%d 应该报错", digits)
		}
	}
	for digits := 0; digits <= 10; digits++ {
		got, err := Format("3.14159265358979", digits, false)
		if err != nil {
			t.Fatalf("digits=%d 不应报错：%v", digits, err)
		}
		frac := ""
		if dot := strings.IndexByte(got, '.'); dot >= 0 {
			frac = got[dot+1:]
		}
		if len(frac) != digits {
			t.Fatalf("digits=%d 输出小数位数不对：%q", digits, got)
		}
	}
}

func TestFormatTrimsSpaces(t *testing.T) {
	checkFormat(t, "  7  ", 1, false, "7.0")
}

func TestFormatMatchesReference(t *testing.T) {
	samples := []string{
		"2.675", "2.5", "3.5", "0.125", "-0.001", "1234567.891",
		"0", "-0", "1", "999.5", "0.5", "1.005", "1234567890.12345",
		".5", "5.", "+7.25", "-9999.9999", "0.0000000001",
	}
	for _, value := range samples {
		for digits := 0; digits <= 10; digits++ {
			for _, group := range []bool{false, true} {
				got, gotErr := Format(value, digits, group)
				want, wantErr := referenceFormat(value, digits, group)
				if (gotErr == nil) != (wantErr == nil) {
					t.Fatalf("Format(%q, %d, %v) 错误状态不一致：%v vs %v", value, digits, group, gotErr, wantErr)
				}
				if got != want {
					t.Fatalf("Format(%q, %d, %v) = %q，参考实现是 %q", value, digits, group, got, want)
				}
			}
		}
	}
}

func TestFormatMatchesReferenceOnRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(20260927))
	for i := 0; i < 10000; i++ {
		var value strings.Builder
		switch rng.Intn(4) {
		case 0:
			value.WriteByte('-')
		case 1:
			value.WriteByte('+')
		}
		intLen := rng.Intn(12)
		for j := 0; j < intLen; j++ {
			value.WriteByte(byte('0' + rng.Intn(10)))
		}
		if rng.Intn(4) != 0 {
			value.WriteByte('.')
			fracLen := rng.Intn(15)
			for j := 0; j < fracLen; j++ {
				value.WriteByte(byte('0' + rng.Intn(10)))
			}
		}
		text := value.String()
		if text == "" || text == "." || text == "-" || text == "+" {
			text = "0"
		}
		digits := rng.Intn(11)
		group := rng.Intn(2) == 0
		got, gotErr := Format(text, digits, group)
		want, wantErr := referenceFormat(text, digits, group)
		if (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("第 %d 个 Format(%q, %d, %v) 错误状态不一致：%v vs %v", i, text, digits, group, gotErr, wantErr)
		}
		if got != want {
			t.Fatalf("第 %d 个 Format(%q, %d, %v) = %q，参考实现是 %q", i, text, digits, group, got, want)
		}
	}
}

func TestFormatLongInput(t *testing.T) {
	// 1 万位输入也不许退化：构造 9999…9.999…5，舍入后应进位成 1000…0.00。
	intPart := strings.Repeat("9", 9990)
	fracPart := strings.Repeat("9", 9) + "5"
	got, err := Format(intPart+"."+fracPart, 2, false)
	if err != nil {
		t.Fatalf("长输入报错：%v", err)
	}
	want := "1" + strings.Repeat("0", 9990) + ".00"
	if got != want {
		t.Fatalf("长输入舍入不对：前 40 位 %q…", got[:40])
	}
	if _, err := Format(intPart+"."+fracPart, 2, true); err != nil {
		t.Fatalf("长输入千分位报错：%v", err)
	}
}

func BenchmarkFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Format("1234567.891", 2, true); err != nil {
			b.Fatal(err)
		}
	}
}
