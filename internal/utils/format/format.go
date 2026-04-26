package format

import "fmt"

func Memory(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
	)
	if bytes >= MB {
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	}
	return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
}
