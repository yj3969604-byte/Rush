package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"log"
	"time"
)

func AwardUser(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	var reqData pojo.AwardInfo
	err := ctx.BindJSON(&reqData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	ip := utils.GetIPAddress(ctx)
	if !utils.InStrings(utils.CsConfig.AwardIps, ip) {
		utils.ErrorBack(ctx, fmt.Sprintf("not in white ip:%s", ip))
		return
	}
	result, err := repository.AwardUser(db, reqData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func getAwardUserHistory(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	var reqData pojo.AwardInfo
	err := ctx.BindJSON(&reqData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	ip := utils.GetIPAddress(ctx)
	if !utils.InStrings(utils.CsConfig.AwardIps, ip) {
		utils.ErrorBack(ctx, fmt.Sprintf("not in white ip:%s", ip))
		return
	}
	result, err := repository.AwardUser(db, reqData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AdminAwardUser godoc
//
//	@Summary		用户余额管理(送钱/扣钱)
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AdminAwardInfo	true	"用户余额管理"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/user/award [post]
func AdminAwardUser(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	var reqData pojo.AdminAwardInfo
	err = ctx.BindJSON(&reqData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := repository.AdminAwardInfo(currentUser, reqData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// UnBindGAuth godoc
//
//	@Summary		解绑谷歌验证
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}		string
//	@Router			/api/v1/outside/user/unbind/gauth [get]
func UnBindGAuth(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	result, err := repository.UnBindGAuth(db, currentUser)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// ChangePass godoc
//
//	@Summary		修改密码
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.UserAdd	true	"修改密码"
//	@Success		200	{object}		pojo.UserBack
//	@Router			/api/v1/outside/user/pass [put]
func ChangePass(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	var userAdd pojo.UserAdd
	err = ctx.ShouldBindJSON(&userAdd)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.ChangePass(db, tempHostInfo, currentUser, userAdd)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetUsers(ctx *gin.Context) {
	var userSearch pojo.UserSearch
	userSearch.SetPageDefaults()
	err := ctx.BindJSON(&userSearch)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)

	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := repository.GetUsers(db, userSearch, currentUser.Username, currentUser.ID)
	utils.SuccessObjBack(ctx, result)
}

func SetUser(ctx *gin.Context) {
	var userAdd pojo.UserAdd
	err := ctx.ShouldBindJSON(&userAdd)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)

	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := repository.SetUser(db, tempHostInfo, userAdd, currentUser.ID)

	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetRoutes(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetRoutes(db, tempHostInfo, currentUser)
	utils.SuccessObjBack(ctx, result)
}

// WhiteUserLogin 白名单测试登录
func WhiteUserLogin(ctx *gin.Context) {
	ip := utils.GetIPAddress(ctx)
	if !utils.InWhiteIps(ip) {
		utils.ErrorBack(ctx, fmt.Sprintf("非法ip:%s", ip))
		return
	}
	var reqUserLogin pojo.UserLogin
	err := ctx.BindJSON(&reqUserLogin)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	onlineUser := pojo.OnlineUser{
		Username:  reqUserLogin.Username,
		Browser:   ctx.GetHeader("User-Agent"),
		Ip:        utils.GetIPAddress(ctx),
		LoginTime: time.Now(),
	}
	db := ctx.MustGet("db").(*gorm.DB)
	data, err := repository.WhiteUserLogin(db, tempHostInfo, reqUserLogin, onlineUser)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

func UserLogin(ctx *gin.Context) {
	var reqUserLogin pojo.UserLogin
	err := ctx.BindJSON(&reqUserLogin)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	password, err := utils.DecPriKey(reqUserLogin.Password, tempHostInfo.PriKey)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	reqUserLogin.Password = password
	onlineUser := pojo.OnlineUser{
		Username:  reqUserLogin.Username,
		Browser:   ctx.GetHeader("User-Agent"),
		Ip:        utils.GetIPAddress(ctx),
		LoginTime: time.Now(),
	}
	db := ctx.MustGet("db").(*gorm.DB)
	data, err := repository.UserLogin(db, tempHostInfo, reqUserLogin, onlineUser)
	if err != nil {
		if err.Error() == "请先绑定二维码" {
			utils.ErrorObjBack(ctx, data, err.Error())
			return
		}
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

func DelUsers(ctx *gin.Context) {
	var req pojo.Ids
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}

	if len(req.Ids) == 0 {
		utils.ErrorBack(ctx, "ids 不能为空")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result, err := repository.DelUsers(db, req.Ids)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// UserCashHistory godoc
//
//	@Summary		金额变动记录分页
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.CashHistorySearch	true	"金额变动记录分页"
//	@Success		200	{object}		pojo.CashHistoryPage
//	@Router			/api/v1/manager/cashHistory [post]
func UserCashHistory(ctx *gin.Context) {
	var req pojo.CashHistorySearch
	req.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if req.UserId == 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	res, err := repository.UserAwardInfos(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, res)
}

// CurrentUserInfo godoc
//
//	@Summary		当前用户信息
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}		pojo.UserBack
//	@Router			/api/v1/manager/userInfo [post]
func CurrentUserInfo(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(int64)
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	user := utils.GetTempUser(tempHostInfo.TablePrefix, userId)
	var tempUserBack pojo.UserBack
	_ = copier.Copy(&tempUserBack, &user)
	log.Printf("tempUserBack: %v", user.RoleStr)
	err := json.Unmarshal([]byte(user.RoleStr), &tempUserBack.Roles)
	if err != nil {
		log.Printf("err: %v", err)
		utils.ErrorBack(ctx, err.Error())
	}
	log.Printf("tempUserBack.Roles:%v", tempUserBack.Roles)
	utils.SuccessObjBack(ctx, tempUserBack)
}

// ResetPassword godoc
//
//	@Summary		重置密码
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Param			data body		pojo.UserResetPwd	true	"重置密码"
//	@Router			/api/v1/manager/resetPassword [post]
func ResetPassword(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(int64)
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	user := utils.GetTempUser(tempHostInfo.TablePrefix, userId)
	var req pojo.UserResetPwd
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	encPassword := utils.EncodePass(tempHostInfo.Salt, req.OldPassword)
	if req.OldPassword != encPassword {
		utils.ErrorBack(ctx, "旧密码输入错误")
	}
	if req.NewPassword == "" {
		utils.ErrorBack(ctx, "请输入有效密码")
	}
	db := ctx.MustGet("db").(*gorm.DB)
	err := repository.ResetPwd(db, req.NewPassword, user.ID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
	}
	utils.SuccessObjBack(ctx, req)
}
