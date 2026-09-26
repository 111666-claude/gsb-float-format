package numfmt

import (
	"strings"
	"testing"
)

func TestFormatIntegerWithDigits(t *testing.T) {
	got, err := Format("1", 2, false)
	if err != nil || got != "1.00" {
		t.Fatalf("格式化不对：%q %v", got, err)
	}
}

func TestFormatDecimalRounding(t *testing.T) {
	cases := map[string]string{
		"2.675":   "2.68",
		"2.674":   "2.67",
		"0.125":   "0.12",
		"0.135":   "0.14",
		"1.005":   "1.00",
		"9.999":   "10.00",
		"0.995":   "1.00",
		"2.6751":  "2.68",
		"0.045":   "0.04",
		"99.9999": "100.00",
	}
	for in, want := range cases {
		got, err := Format(in, 2, false)
		if err != nil || got != want {
			t.Errorf("Format(%q, 2) = %q, %v；期望 %q", in, got, err, want)
		}
	}
}

func TestFormatRoundsHalfToEven(t *testing.T) {
	cases := []struct {
		in     string
		digits int
		want   string
	}{
		{"2.5", 0, "2"},
		{"3.5", 0, "4"},
		{"0.5", 0, "0"},
		{"1.5", 0, "2"},
		{"4.5", 0, "4"},
		{"0.125", 2, "0.12"},
		{"0.375", 2, "0.38"},
		{"2.685", 2, "2.68"},
		{"2.695", 2, "2.70"},
		{"10.5", 0, "10"},
		{"11.5", 0, "12"},
	}
	for _, c := range cases {
		got, err := Format(c.in, c.digits, false)
		if err != nil || got != c.want {
			t.Errorf("Format(%q, %d) = %q, %v；期望 %q", c.in, c.digits, got, err, c.want)
		}
	}
}

func TestFormatKeepsTrailingZeros(t *testing.T) {
	cases := []struct {
		in     string
		digits int
		want   string
	}{
		{"1.2", 3, "1.200"},
		{"2.695", 2, "2.70"},
		{"7", 0, "7"},
		{"7", 4, "7.0000"},
		{"0.1", 10, "0.1000000000"},
	}
	for _, c := range cases {
		got, err := Format(c.in, c.digits, false)
		if err != nil || got != c.want {
			t.Errorf("Format(%q, %d) = %q, %v；期望 %q", c.in, c.digits, got, err, c.want)
		}
	}
}

func TestFormatGrouping(t *testing.T) {
	cases := []struct {
		in     string
		digits int
		want   string
	}{
		{"1234567.891", 2, "1,234,567.89"},
		{"-1234567.891", 2, "-1,234,567.89"},
		{"1000", 0, "1,000"},
		{"100", 0, "100"},
		{"12345678901234", 0, "12,345,678,901,234"},
		{"999.999", 2, "1,000.00"},
		{"-0.5", 0, "0"},
	}
	for _, c := range cases {
		got, err := Format(c.in, c.digits, true)
		if err != nil || got != c.want {
			t.Errorf("Format(%q, %d, group) = %q, %v；期望 %q", c.in, c.digits, got, err, c.want)
		}
	}
}

func TestFormatNegativeZero(t *testing.T) {
	cases := []struct {
		in     string
		digits int
		want   string
	}{
		{"-0.001", 2, "0.00"},
		{"-0", 2, "0.00"},
		{"-0.0", 0, "0"},
		{"-0.004", 2, "0.00"},
		{"-0.005", 2, "0.00"},
		{"-0.1", 0, "0"},
		{"-0.5", 0, "0"},
		{"-0.006", 2, "-0.01"},
	}
	for _, c := range cases {
		got, err := Format(c.in, c.digits, false)
		if err != nil || got != c.want {
			t.Errorf("Format(%q, %d) = %q, %v；期望 %q", c.in, c.digits, got, err, c.want)
		}
	}
}

func TestFormatRejectsGarbage(t *testing.T) {
	for _, in := range []string{"abc", "", "1.2.3", "--1", "1e5", "0x10", ".", "-", "1,000"} {
		if _, err := Format(in, 2, false); err == nil {
			t.Errorf("非数值 %q 应该报错", in)
		}
	}
}

func TestFormatErrorCarriesInput(t *testing.T) {
	if _, err := Format("abc", 2, false); err == nil || !strings.Contains(err.Error(), "abc") {
		t.Fatalf("错误信息要带原串：%v", err)
	}
}

func TestFormatDigitsRange(t *testing.T) {
	for digits := 0; digits <= 10; digits++ {
		got, err := Format("3.14159265358979", digits, false)
		if err != nil {
			t.Fatalf("digits=%d 不应报错：%v", digits, err)
		}
		frac := ""
		if i := strings.IndexByte(got, '.'); i >= 0 {
			frac = got[i+1:]
		}
		if len(frac) != digits {
			t.Fatalf("digits=%d 输出小数位应为 %d：%q", digits, digits, got)
		}
	}
	for _, digits := range []int{-1, 11} {
		if _, err := Format("1", digits, false); err == nil {
			t.Errorf("digits=%d 应报错", digits)
		}
	}
}

func TestFormatDeterministic(t *testing.T) {
	first, err := Format("2.675", 2, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		got, err := Format("2.675", 2, true)
		if err != nil || got != first {
			t.Fatalf("重复格式化不一致：%q vs %q", first, got)
		}
	}
}

func TestFormatTrimsSpaces(t *testing.T) {
	got, err := Format("  7  ", 1, false)
	if err != nil || got != "7.0" {
		t.Fatalf("带空白的输入不对：%q %v", got, err)
	}
}

func TestFormatLongInput(t *testing.T) {
	integer := strings.Repeat("9", 5000)
	frac := strings.Repeat("9", 4999) + "5"
	got, err := Format(integer+"."+frac, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	want := "1" + strings.Repeat("0", 5000)
	grouped := groupThousands(want) + ".00"
	if got != grouped {
		t.Fatalf("长输入进位不对：got 长度 %d", len(got))
	}
}
