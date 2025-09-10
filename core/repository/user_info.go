package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
	"log"
	"time"
)

func AdminAwardInfo(currentUser pojo.SysUser, reqData pojo.AdminAwardInfo) (result string, err error) {
	awardUni := utils.RandomString(6)
	checkKey := fmt.Sprintf(utils.KeyUserAwardCheck, reqData.UserId, awardUni)
	checkValue := utils.MD5(checkKey + "nj")
	awardInfo := pojo.AwardInfo{
		CheckKey:   checkKey,
		CheckValue: checkValue,
	}
	awardInfo.AwardUnis = append(awardInfo.AwardUnis, pojo.AwardUni{
		UserId:     reqData.UserId,
		Amount:     reqData.Amount,
		AwardUni:   awardUni,
		CashMark:   reqData.CashMark,
		CashDesc:   "",
		RefuseCash: false,
		FromUserId: currentUser.ID,
	})
	utils.RD.SetEX(context.Background(), checkKey, checkValue, 1*time.Minute)
	requestData, _ := json.Marshal(awardInfo)
	response, _, err := utils.ProxyPostRequest(utils.CsConfig.AwardUrl, utils.JsonHead, requestData, nil)
	log.Printf("AdminAwardInfo response = %s", string(response))
	if err == nil {
		var responseObj pojo.BaseResponse
		_ = json.Unmarshal(response, &responseObj)
		if responseObj.Success {
			return responseObj.Message, err
		}
	}
	return result, err
}

func LocalAwardInfo(currentUser pojo.SysUser, reqData pojo.AwardInfo) (result string, err error) {
	if len(reqData.AwardUnis) == 0 {
		return result, errors.New("data_error")
	}
	awardUni := reqData.AwardUnis[0].AwardUni
	reqData.CheckKey = fmt.Sprintf(utils.KeyUserAwardCheck, currentUser.ID, awardUni)
	reqData.CheckValue = utils.MD5(reqData.CheckKey + "nj")
	utils.RD.SetEX(context.Background(), reqData.CheckKey, reqData.CheckValue, 1*time.Minute)
	requestData, _ := json.Marshal(reqData)
	response, _, err := utils.ProxyPostRequest(utils.CsConfig.AwardUrl, utils.JsonHead, requestData, nil)
	log.Printf("LocalAwardInfo response = %s", string(response))
	if err == nil {
		var responseObj pojo.BaseResponse
		_ = json.Unmarshal(response, &responseObj)
		if responseObj.Success {
			return responseObj.Message, err
		}
	}
	return result, err
}

func AwardUser(db *gorm.DB, reqData pojo.AwardInfo) (result []int64, err error) {
	checkValue := utils.GetRdString(reqData.CheckKey, "")
	if reqData.CheckValue != checkValue {
		return result, errors.New(fmt.Sprintf("AwardUser error request not check %s %s->%s", reqData.CheckKey, checkValue, reqData.CheckValue))
	}
	if len(reqData.AwardUnis) == 0 {
		return result, errors.New("data error")
	}
	lockKeys := make([]string, 0)
	updateUsers := make([]string, 0)
	tx := db.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Printf("db begin error.err=%v", p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
			for _, updateUser := range updateUsers {
				_ = utils.PublishMQ(utils.MQMessage{
					MessageType: utils.KeyMqUserUpdate,
					Data:        updateUser,
				})
			}
		}
		for _, lockKey := range lockKeys {
			_ = utils.ReleaseLock(lockKey)
			//log.Printf("释放锁:%s", lockKey)
		}
	}()
	for _, awardUni := range reqData.AwardUnis {
		if awardUni.UserId == 0 || awardUni.Amount == 0 || awardUni.AwardUni == "" {
			return result, errors.New("AwardUser error data not check")
		}
		prefix := utils.GetDbPrefix(db)
		awardUser := utils.GetTempUser(prefix, awardUni.UserId)
		blockKey := fmt.Sprintf(utils.KeyLockUserAward, awardUser.ID)
		acquired := false
		for i := 0; i < 100; i++ { // 重试10s
			acquired, err = utils.AcquireLock(blockKey, 20*time.Second)
			if acquired {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if !acquired {
			fmt.Printf("AwardUser error acquiring lock for user %s: %v\n", awardUser.Username, err)
			err = errors.New("block user error")
			return result, err
		}
		//log.Printf("加锁:%s", blockKey)
		lockKeys = append(lockKeys, blockKey)
		var currentUser pojo.SysUser
		tx.Where("id = ?", awardUser.ID).First(&currentUser)
		if awardUni.Amount < 0 && awardUser.Amount+awardUni.Amount < 0 {
			err = errors.New(fmt.Sprintf("amount2 not enough.%.3f->%.3f", awardUser.Amount, awardUni.Amount)) // 余额不足
			return result, err
		}
		cashHistory := pojo.CashHistory{
			UserId:      currentUser.ID,
			AwardUni:    awardUni.AwardUni,
			Amount:      awardUni.Amount,
			StartAmount: currentUser.Amount,
			EndAmount:   utils.ToMoney(currentUser.Amount).Add(utils.ToMoney(awardUni.Amount)).ToDollars(),
			CashMark:    awardUni.CashMark,
			CashDesc:    awardUni.CashDesc,
			FromUserId:  awardUni.FromUserId,
		}
		err = tx.Create(&cashHistory).Error
		if err != nil {
			return result, err
		}
		update := make(map[string]any)
		update["amount"] = gorm.Expr(fmt.Sprintf("amount + %.3f", awardUni.Amount))
		if awardUni.Amount > 0 && !awardUni.RefuseCash {
			update["top_amount"] = gorm.Expr(fmt.Sprintf("top_amount + %.3f", awardUni.Amount))
		}
		err = tx.Model(&awardUser).Updates(update).Error
		if err != nil {
			return result, err
		}
		updateUsers = append(updateUsers, fmt.Sprintf("%d#%s", awardUser.ID, prefix))
		result = append(result, cashHistory.ID)
	}
	return result, err
}

func GetUsers(db *gorm.DB, userSearch pojo.UserSearch, currentUserName string, currentUserId int64) (result pojo.UserResp) {
	var users []pojo.SysUser
	//db = db.Model(&pojo.SysUser{}).Where("user_type != ?", 3)
	db = db.Model(&pojo.SysUser{})
	if userSearch.Username != "" {
		db = db.Where("username like ?", "%"+userSearch.Username+"%")
	}
	if currentUserName != "admin" {
		db.Where("parent_id = ?", currentUserId)
	}
	if userSearch.Enabled != nil {
		db = db.Where("enabled = ?", userSearch.Enabled)
	}
	db.Model(&pojo.SysUser{}).Count(&result.Total)
	db = db.Order("id desc").Limit(userSearch.PageSize).Offset(userSearch.PageSize * userSearch.CurrentPage)
	db.Find(&users)
	for _, user := range users {
		var tempUserBack pojo.UserBack
		_ = copier.Copy(&tempUserBack, &user)
		log.Printf("tempUserBack: %v", user.RoleStr)
		err := json.Unmarshal([]byte(user.RoleStr), &tempUserBack.Roles)
		if err != nil {
			log.Printf("err: %v", err)
		}
		log.Printf("tempUserBack.Roles:%v", tempUserBack.Roles)
		result.List = append(result.List, tempUserBack)
	}
	result.PageSize = userSearch.PageSize
	result.CurrentPage = userSearch.CurrentPage
	return result
}

func buildMenuTree(menus []pojo.SysMenu, parentID int64) []pojo.BackMenu {
	var result []pojo.BackMenu
	for _, menu := range menus {
		var tempMenu pojo.BackMenu
		_ = copier.Copy(&tempMenu, &menu)
		if tempMenu.ParentID == parentID {
			tempMenu.Children = buildMenuTree(menus, tempMenu.ID)
			_ = json.Unmarshal([]byte(menu.MetaStr), &tempMenu.Meta)
			result = append(result, tempMenu)
		}
	}
	return result
}

func UnBindGAuth(db *gorm.DB, currentUser pojo.SysUser) (result pojo.UserBack, err error) {
	currentUser.BindCode = false
	currentUser.GoogleCode = ""
	err = db.Save(&currentUser).Error
	_ = copier.Copy(&result, &currentUser)
	_ = json.Unmarshal([]byte(currentUser.RoleStr), &result.Roles)
	return result, err
}

func ChangePass(db *gorm.DB, hostInfo pojo.HostInfo, currentUser pojo.SysUser, userAdd pojo.UserAdd) (result pojo.UserBack, err error) {
	isManager := currentUser.UserType == 1 || currentUser.UserType == 2
	if !isManager {
		userAdd.ID = currentUser.ID
	}
	dbUser := utils.GetTempUser(hostInfo.TablePrefix, userAdd.ID)
	if dbUser.ID == 0 {
		return result, errors.New("操作的数据不存在")
	}
	if len(userAdd.Password) < 6 || len(userAdd.Password) > 18 {
		return result, errors.New("密码长度必须在 6-18 位以内")
	}
	dbUser.Password = utils.EncodePass(hostInfo.Salt, userAdd.Password)
	err = db.Save(&dbUser).Error
	_ = copier.Copy(&result, &dbUser)
	_ = json.Unmarshal([]byte(dbUser.RoleStr), &result.Roles)
	return result, err
}

func SetUser(db *gorm.DB, hostInfo pojo.HostInfo, userAdd pojo.UserAdd, currentUserId int64) (result pojo.UserBack, err error) {
	dbUser := utils.GetTempUser(hostInfo.TablePrefix, userAdd.ID)
	if dbUser.ID == 0 {
		db.Where("username = ?", userAdd.Username).First(&dbUser)
		if dbUser.ID != 0 {
			return result, errors.New("用户名重复")
		}
		_ = copier.Copy(&dbUser, &userAdd)
		roleStr, _ := json.Marshal(userAdd.Roles)
		dbUser.UniKey = utils.GetUserUniKey(hostInfo.TablePrefix)
		dbUser.SecurityKey = utils.RandomString(32)
		dbUser.RoleStr = string(roleStr)
		dbUser.Password = utils.EncodePass(hostInfo.Salt, userAdd.Password)
		err = db.Create(&dbUser).Error
	} else {
		if dbUser.Username != userAdd.Username {
			return result, errors.New("用户名不可修改")
		}
		userAdd.Password = ""
		_ = copier.Copy(&dbUser, userAdd)
		roleStr, _ := json.Marshal(userAdd.Roles)
		dbUser.RoleStr = string(roleStr)
		dbUser.Enabled = userAdd.Enabled
		dbUser.Mark = userAdd.Mark
		err = db.Save(&dbUser).Error
	}
	_ = copier.Copy(&result, &dbUser)
	_ = json.Unmarshal([]byte(dbUser.RoleStr), &result.Roles)
	utils.UpdateTempUser(hostInfo.TablePrefix, dbUser)
	return result, err
}

func GetRoutes(db *gorm.DB, hostInfo pojo.HostInfo, currentUser pojo.SysUser) (result []pojo.BackMenu) {
	_ = json.Unmarshal([]byte(currentUser.RoleStr), &currentUser.Roles)
	var menus []pojo.SysMenu
	if currentUser.UserType == 1 {
		db.Find(&menus)
	} else {
		currentUserStr, _ := json.Marshal(currentUser)
		log.Printf("currentUser=%s", string(currentUserStr))
		var roles []pojo.SysRole
		db.Where("code in ?", currentUser.Roles).Find(&roles)
		menuIds := make([]int64, 0)
		for _, role := range roles {
			var tempIds []int64
			_ = json.Unmarshal([]byte(role.MenuIdStr), &tempIds)
			for _, tempId := range tempIds {
				if utils.InInt64s(menuIds, tempId) {
					continue
				}
				menuIds = append(menuIds, tempId)
			}
		}
		db.Where("id in ?", menuIds).Find(&menus)
	}
	//menusStr, _ := json.Marshal(menus)
	//log.Printf("menus=%s", string(menusStr))
	endMenus := make([]pojo.SysMenu, 0)
	for _, menu := range menus {
		if menu.Name == "SystemHostInfo" && hostInfo.HostName != utils.CsConfig.DefaultHost.HostName {
			continue
		}
		endMenus = append(endMenus, menu)
	}
	return buildMenuTree(endMenus, 0)
}

func WhiteUserLogin(db *gorm.DB, hostInfo pojo.HostInfo, reqUserLogin pojo.UserLogin, onlineUser pojo.OnlineUser) (data pojo.LoginBack, err error) {
	reqUserLoginStr, _ := json.Marshal(reqUserLogin)
	log.Printf("userLogin=%s;host=%s", string(reqUserLoginStr), hostInfo.HostName)
	var dbUser *pojo.SysUser
	db.Where("username = ?", reqUserLogin.Username).First(&dbUser)
	dbUserStr, _ := json.Marshal(dbUser)
	log.Printf("dbUser=%s", string(dbUserStr))
	if dbUser.ID == 0 {
		return data, errors.New("user login error")
	}
	if utils.CheckPasswordHash(reqUserLogin.Password, dbUser.Password, utils.GlobalConfig.Salt) {
		log.Printf("user login error.pass error userId = %d,pass=%s", dbUser.ID, reqUserLogin.Password)
		return data, errors.New("user login error")
	}
	if dbUser.UserType != 1 && dbUser.UserType != 2 && dbUser.UserType != 3 {
		return data, errors.New("user error")
	}
	if !dbUser.Enabled {
		return data, errors.New("User account disabled")
	}
	data = GetUserInfo(*dbUser)
	token, err := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, dbUser.Username, dbUser.ID, dbUser.UserType, hostInfo.HostName)
	data.AccessToken = token
	key := utils.KeyRdOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Key = key
	//log.Printf("onlineUser=%v", onlineUser)
	userJSON, _ := json.Marshal(onlineUser)
	//log.Printf("userJSON=%v", string(userJSON))
	utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second)
	return data, err
}

func UserLogin(db *gorm.DB, hostInfo pojo.HostInfo, reqUserLogin pojo.UserLogin, onlineUser pojo.OnlineUser) (data pojo.LoginBack, err error) {
	reqUserLoginStr, _ := json.Marshal(reqUserLogin)
	log.Printf("userLogin=%s;host=%s", string(reqUserLoginStr), hostInfo.HostName)
	var dbUser *pojo.SysUser
	db.Where("username = ?", reqUserLogin.Username).First(&dbUser)
	dbUserStr, _ := json.Marshal(dbUser)
	log.Printf("dbUser=%s", string(dbUserStr))
	if dbUser.ID == 0 {
		return data, errors.New("user login error")
	}
	needBind := false
	if dbUser.GoogleCode != "" {
		if reqUserLogin.Code == "" {
			return data, errors.New("请输入验证码")
		}
		_, err2 := totp.Generate(totp.GenerateOpts{
			Issuer:      "gg",
			AccountName: dbUser.Username,
			Secret:      []byte(dbUser.GoogleCode),
		})
		if err2 != nil {
			return data, err2
		}
		valid := totp.Validate(reqUserLogin.Code, dbUser.GoogleCode)
		if !valid {
			if dbUser.BindCode {
				return data, errors.New("验证码错误")
			}
			//needBind = true
		} else {
			db.Model(&dbUser).Update("bind_code", true)
			utils.UpdateTempUser(hostInfo.TablePrefix, *dbUser)
		}
	} else {
		needBind = true
	}
	if needBind {
		key, err2 := totp.Generate(totp.GenerateOpts{
			Issuer:      "sg",
			AccountName: dbUser.Username,
		})
		if err2 != nil {
			return data, err2
		}
		db.Model(&dbUser).Update("google_code", key.Secret())
		utils.UpdateTempUser(hostInfo.TablePrefix, *dbUser)
		qrCode, err2 := qrcode.New(key.URL(), qrcode.Medium)
		if err2 != nil {
			return data, err2
		}
		pngData, err2 := qrCode.PNG(200)
		if err2 != nil {
			return data, err2
		}
		data.QrCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
		return data, errors.New("请先绑定二维码")
	}
	if utils.CheckPasswordHash(reqUserLogin.Password, dbUser.Password, utils.GlobalConfig.Salt) {
		log.Printf("user login error.pass error userId = %d,pass=%s", dbUser.ID, reqUserLogin.Password)
		return data, errors.New("user login error")
	}
	if dbUser.UserType != 1 && dbUser.UserType != 2 && dbUser.UserType != 3 {
		return data, errors.New("user error")
	}
	if !dbUser.Enabled {
		return data, errors.New("User account disabled")
	}
	data = GetUserInfo(*dbUser)
	token, err := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, dbUser.Username, dbUser.ID, dbUser.UserType, hostInfo.HostName)
	data.AccessToken = token
	key := utils.KeyRdOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Key = key
	//log.Printf("onlineUser=%v", onlineUser)
	userJSON, _ := json.Marshal(onlineUser)
	//log.Printf("userJSON=%v", string(userJSON))
	utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second)
	return data, err
}

func GetUserInfo(dbUser pojo.SysUser) (data pojo.LoginBack) {
	userBak := pojo.UserBack{}
	_ = copier.Copy(&userBak, dbUser)
	data.Username = userBak.Username
	data.UserType = userBak.UserType
	data.Roles = dbUser.Roles
	return data
}

func DelUsers(db *gorm.DB, ids []int64) (result string, err error) {
	var users []pojo.SysUser
	if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return result, err
	}
	for _, user := range users {
		if user.Username == "admin" {
			return result, errors.New("admin用户不可删除")
		}
	}
	if err := db.Where("id IN ?", ids).Delete(&pojo.SysUser{}).Error; err != nil {
		return result, err
	}
	return result, nil
}

func UserAwardInfos(db *gorm.DB, search pojo.CashHistorySearch) (result pojo.CashHistoryPage, err error) {
	var cashHistoryList []pojo.CashHistory
	db = db.Model(&pojo.CashHistory{}).Where("user_id = ?", search.UserId)

	db.Model(&pojo.CashHistory{}).Count(&result.Total)
	db = db.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	db.Find(&cashHistoryList)

	for _, user := range cashHistoryList {
		var tempUserBack pojo.CashHistoryResp
		_ = copier.Copy(&tempUserBack, &user)
		result.List = append(result.List, tempUserBack)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage

	return result, err
}

func ResetPwd(db *gorm.DB, newPassword string, userId int64) (err error) {
	result := db.Model(&pojo.SysUser{}).
		Where("id = ?", userId).
		Update("password", newPassword)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", userId)
	}

	return nil
}
