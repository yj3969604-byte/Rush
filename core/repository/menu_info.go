package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func DelMenu(db *gorm.DB, currentUser pojo.SysUser, id string) (result string, err error) {
	var dbMenu pojo.SysMenu
	db.Where("id = ?", id).Find(&dbMenu)
	if dbMenu.ID == 0 {
		return result, errors.New("删除的菜单不存在")
	}
	_ = json.Unmarshal([]byte(currentUser.RoleStr), &currentUser.Roles)
	var roles []pojo.SysRole
	db.Where("id in ?", currentUser.Roles).Find(&roles)
	haveRole := currentUser.Username == "admin"
	if !haveRole {
		for _, role := range roles {
			var menuIds []string
			_ = json.Unmarshal([]byte(role.MenuIdStr), &menuIds)
			if utils.InStrings(menuIds, id) {
				haveRole = true
				break
			}
		}
	}
	if !haveRole {
		return result, errors.New("没有删除该菜单的权限")
	}
	db.Delete(&dbMenu)
	return "success", nil
}

func SetMenus(db *gorm.DB, menuSet pojo.MenuSet) (result string, err error) {
	var dbMenu pojo.SysMenu
	if menuSet.ID > 0 {
		db.Where("id = ?", menuSet.ID).Find(&dbMenu)
		if dbMenu.ID == 0 {
			return result, errors.New("更新的菜单不存在")
		}
		_ = copier.Copy(&dbMenu, menuSet)
		_ = copier.Copy(&dbMenu.Meta, menuSet)
		_ = copier.Copy(&dbMenu.Meta.Transition, menuSet)
		if menuSet.MenuType != dbMenu.MenuType {
			dbMenu.MenuType = menuSet.MenuType
		}
		metaStr, _ := json.Marshal(dbMenu.Meta)
		dbMenu.MetaStr = string(metaStr)
		db.Save(&dbMenu)
	} else {
		_ = copier.Copy(&dbMenu, &menuSet)
		_ = copier.Copy(&dbMenu.Meta, &menuSet)
		_ = copier.Copy(&dbMenu.Meta.Transition, &menuSet)
		if menuSet.MenuType != dbMenu.MenuType {
			dbMenu.MenuType = menuSet.MenuType
		}
		metaStr, _ := json.Marshal(dbMenu.Meta)
		dbMenu.MetaStr = string(metaStr)
		db.Create(&dbMenu)
	}
	return "success", nil
}

func GetMenus(db *gorm.DB, hostInfo pojo.HostInfo) (result []pojo.MenuSet) {
	var menus []pojo.SysMenu
	db.Find(&menus)
	result = make([]pojo.MenuSet, 0)
	for _, menu := range menus {
		if hostInfo.HostName != utils.CsConfig.DefaultHost.HostName && menu.Name == "SystemHostInfo" {
			continue
		}
		_ = json.Unmarshal([]byte(menu.MetaStr), &menu.Meta)
		var tempMenuSet pojo.MenuSet
		_ = copier.Copy(&tempMenuSet, &menu)
		_ = copier.Copy(&tempMenuSet, &menu.Meta)
		_ = copier.Copy(&tempMenuSet, &menu.Meta.Transition)
		result = append(result, tempMenuSet)
	}
	return result
}
