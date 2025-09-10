package utils

import (
	"log"
	"math/rand/v2"
)

type WidthInfo[T any] struct {
	Width int `json:"width"` // 对应权重
	Data  T   `json:"data"`  // 对应原始数据
}

func GetWidthData[T any](datas []WidthInfo[T]) (result *T) {
	if len(datas) == 0 {
		return result
	}
	// 总权重
	allWidth := 0
	for _, v := range datas {
		if v.Width == 0 {
			// 移除无权重的用户
			continue
		}
		allWidth += v.Width
	}
	// 检查是否有可用的用户和任务
	if allWidth == 0 {
		return &datas[rand.IntN(len(datas))].Data
	}
	endWidth := rand.IntN(allWidth)
	for _, v := range datas {
		if v.Width == 0 {
			continue
		}
		log.Printf("allWidth=%d;endWidth=%d;userWidth=%d", allWidth, endWidth, v.Width)
		if v.Width < endWidth {
			endWidth = endWidth - v.Width
			continue
		}
		return &v.Data
	}
	return result
}
