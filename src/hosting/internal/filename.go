package internal

import (
	"fmt"
	"strings"
	"time"
)

// GenNewName 生成新文件名,避免重名
func GenNewName(old string) string {
	dotIdx := strings.LastIndexByte(old, '.')
	return fmt.Sprintf("%s-%d%s", old[:dotIdx], time.Now().UnixNano(), old[dotIdx:])
}
