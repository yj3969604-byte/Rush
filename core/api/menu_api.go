package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DelMenu(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	id := ctx.Param("id")
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelMenu(db, currentUser, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetMenus(ctx *gin.Context) {
	var menuSet pojo.MenuSet
	err := ctx.ShouldBindJSON(&menuSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, _ := repository.SetMenus(db, menuSet)
	utils.SuccessObjBack(ctx, result)
}

func GetMenus(ctx *gin.Context) {
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetMenus(db, tempHostInfo)
	utils.SuccessObjBack(ctx, result)
}
