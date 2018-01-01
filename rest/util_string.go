package rest

import (
	"fmt"
	"strconv"
)

//把一个大小转变成方便读的格式
//human readable file size
func HumanFileSize(bytes int64, si bool) string {
	var thresh int64 = 1000
	if si {
		thresh = 1024
	}
	if bytes < 0 {
		bytes = 0
	}
	if bytes < thresh {
		return fmt.Sprintf("%dB", bytes)
	}
	var units = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	if si {
		units = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	}
	var u = 0
	var tmp = float64(bytes)
	var standard = float64(thresh)
	for tmp >= standard && u < len(units)-1 {
		tmp /= float64(standard)
		u++
	}

	numStr := strconv.FormatFloat(tmp, 'f', 1, 64)

	return fmt.Sprintf("%s%s", numStr, units[u])
}
