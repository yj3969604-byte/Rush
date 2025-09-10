package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetRoleMenuIds(ctx *gin.Context) {
	roleId := ctx.Param("roleId")
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetRoleMenuIds(db, roleId)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetRoleMenus(ctx *gin.Context) {
	var roleSearch pojo.RoleSearch
	err := ctx.BindJSON(&roleSearch)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetRoleMenus(db, roleSearch)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetRoleIds(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.GetRoleIds(db, tempHostInfo, userId)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelRole(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	id := ctx.Param("id")
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelRole(db, currentUser, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetRole(ctx *gin.Context) {
	var roleSet pojo.RoleSet
	err := ctx.ShouldBindJSON(&roleSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetRole(db, roleSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetRoles(ctx *gin.Context) {
	var roleSearch pojo.RoleSearch
	err := ctx.BindJSON(&roleSearch)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetRoles(db, roleSearch)
	utils.SuccessObjBack(ctx, result)
}
