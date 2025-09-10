package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
)

func GetHostInfos(ctx *gin.Context) {
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	var hostInfoSearch pojo.HostInfoSearch
	err := ctx.BindJSON(&hostInfoSearch)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := repository.GetHostInfos(tempHostInfo, hostInfoSearch)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetHostInfo(ctx *gin.Context) {
	var hostInfoSet pojo.HostInfoSet
	err := ctx.ShouldBindJSON(&hostInfoSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.SetHostInfo(tempHostInfo, hostInfoSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelHostInfo(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := repository.DelHostInfo(id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
