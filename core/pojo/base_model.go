package pojo

import (
	"time"
)

type BaseModel struct {
	ID        int64     `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

type Ids struct {
	Ids []int64 `json:"ids"`
}

type PageInfo struct {
	CurrentPage int `json:"currentPage" form:"currentPage"`
	PageSize    int `json:"pageSize" form:"pageSize"`
}

type BasePageResponse[T any] struct {
	List        []T   `json:"list"`
	Total       int64 `json:"total"`
	PageSize    int   `json:"pageSize"`
	CurrentPage int   `json:"currentPage"`
}

type BaseObjResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
	Code    int    `json:"code"`
	Success bool   `json:"success"`
}

type BaseResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Code    int    `json:"code"`
	Success bool   `json:"success"`
}

type EncBackData struct {
	Message any `json:"message"`
}

type EncResponse struct {
	Data any `json:"data"`
}
