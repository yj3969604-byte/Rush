package utils

import (
	"BaseGoUni/core/pojo"
	"encoding/json"
	"log"
)

func InitConfig(dbHostInfo pojo.HostInfo) (err error) {
	db := NewPrefixDb(dbHostInfo.TablePrefix)
	var defaultUser pojo.SysUser
	db.Where("username = ?", CsConfig.DefaultUser.Username).First(&defaultUser)
	if defaultUser.ID == 0 {
		roleStr, _ := json.Marshal(CsConfig.DefaultUser.Roles)
		CsConfig.DefaultUser.RoleStr = string(roleStr)
		CsConfig.DefaultUser.Password = EncodePass(dbHostInfo.Salt, CsConfig.DefaultUser.Password)
		db.Create(&CsConfig.DefaultUser)
		userStr, _ := json.Marshal(CsConfig.DefaultUser)
		log.Printf("create user = %s", string(userStr))
	}
	for _, defaultMenu := range CsConfig.DefaultMenus {
		var menu pojo.SysMenu
		db.Where("name = ?", defaultMenu.Name).First(&menu)
		if menu.ID == 0 {
			if defaultMenu.ParentName != "" {
				db.Table(pojo.MenuTableName).Where("name = ?", defaultMenu.ParentName).
					Select("id").Scan(&defaultMenu.ParentID)
			}
			metaStr, _ := json.Marshal(defaultMenu.Meta)
			defaultMenu.MetaStr = string(metaStr)
			db.Create(&defaultMenu)
			menuStr, _ := json.Marshal(defaultMenu)
			log.Printf("create menu = %s", string(menuStr))
		}
	}
	for _, role := range CsConfig.DefaultRoles {
		var defaultRole pojo.SysRole
		db.Where("code = ?", role.Name).First(&defaultRole)
		if defaultRole.ID == 0 {
			menuIds := make([]int64, 0)
			db.Table(pojo.MenuTableName).Where("name in ?", role.MenuNames).Select("id").Scan(&menuIds)
			menuIdStr, _ := json.Marshal(menuIds)
			role.MenuIdStr = string(menuIdStr)
			db.Create(&role)
			roleStr, _ := json.Marshal(role)
			log.Printf("create role = %s", string(roleStr))
		}
	}
	return err
}

func InitMenus(dbHostInfo pojo.HostInfo) (err error) {
	db := NewPrefixDb(dbHostInfo.TablePrefix)
	for _, defaultMenu := range CsConfig.NewMenus {
		var menu pojo.SysMenu
		db.Where("name = ?", defaultMenu.Name).First(&menu)
		if menu.ID == 0 {
			if defaultMenu.ParentName != "" {
				db.Where("name = ?", defaultMenu.ParentName).Select("id").Scan(&defaultMenu.ParentID)
			}
			metaStr, _ := json.Marshal(defaultMenu.Meta)
			defaultMenu.MetaStr = string(metaStr)
			db.Create(&defaultMenu)
			menuStr, _ := json.Marshal(defaultMenu)
			log.Printf("create menu = %s", string(menuStr))
		}
	}
	return err
}
