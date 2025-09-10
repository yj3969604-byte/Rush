package utils

import (
	"BaseGoUni/core/pojo"
)

func ScanHostUni(back func(info pojo.HostInfo)) {
	size := 5000
	page := 0
	count := size
	uniTables := make(map[string]int)
	for ; count == size; page++ {
		size = scanHostUni(back, page, size, &uniTables)
	}
}
func scanHostUni(back func(info pojo.HostInfo), page int, size int, uniTables *map[string]int) int {
	var hostInfos []pojo.HostInfo
	Db.Limit(size).Offset(page * size).Order("id asc").Find(&hostInfos)
	for _, hostInfo := range hostInfos {
		if (*uniTables)[hostInfo.TablePrefix] > 0 {
			continue
		}
		(*uniTables)[hostInfo.TablePrefix] = 1
		back(hostInfo)
	}
	return len(hostInfos)
}
