// Package numfmt 按十进制规则格式化数值。
package numfmt

import (
	"fmt"
	"strconv"
	"strings"
)

// Format 保留 digits 位小数。缺陷：先转 float 再用 Sprintf，二进制舍入、不做千分位、也不消 -0。
func Format(value string, digits int, group bool) (string, error) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return "", fmt.Errorf("不是合法的数值：%q", value)
	}
	_ = group
	return fmt.Sprintf("%.*f", digits, number), nil
}
