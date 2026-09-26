package numfmt

import (
	"strings"
	"testing"
	"time"
)

func TestPerfMillionFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过")
	}
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		if _, err := Format("1234567.895", 2, true); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("100 万次格式化耗时 %v", time.Since(start))
}

func TestPerfLongInput(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过")
	}
	in := strings.Repeat("7", 5000) + "." + strings.Repeat("3", 4999) + "5"
	start := time.Now()
	for i := 0; i < 1000; i++ {
		if _, err := Format(in, 2, true); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("1 万位输入 1000 次耗时 %v", time.Since(start))
}
