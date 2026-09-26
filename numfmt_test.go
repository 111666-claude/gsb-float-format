package numfmt

import "testing"

func TestFormatIntegerWithDigits(t *testing.T) {
got, err := Format("1", 2, false)
if err != nil || got != "1.00" {
t.Fatalf("格式化不对：%q %v", got, err)
}
}

func TestFormatRoundsHalfToEven(t *testing.T) {
got, err := Format("2.5", 0, false)
if err != nil || got != "2" {
t.Fatalf("2.5 按四舍六入五成双应是 2：%q", got)
}
}

func TestFormatKeepsTrailingZeros(t *testing.T) {
got, err := Format("1.2", 3, false)
if err != nil || got != "1.200" {
t.Fatalf("尾随零不对：%q", got)
}
}

func TestFormatRejectsGarbage(t *testing.T) {
if _, err := Format("abc", 2, false); err == nil {
t.Fatal("非数值应该报错")
}
}

func TestFormatTrimsSpaces(t *testing.T) {
got, err := Format("  7  ", 1, false)
if err != nil || got != "7.0" {
t.Fatalf("带空白的输入不对：%q %v", got, err)
}
}
