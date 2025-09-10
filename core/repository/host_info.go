package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"github.com/jinzhu/copier"
)

func GetHostInfos(hostInfo pojo.HostInfo, hostInfoSearch pojo.HostInfoSearch) (result pojo.HostInfoResp, err error) {
	if hostInfo.HostName != utils.CsConfig.DefaultHost.HostName {
		return result, errors.New("域名错误")
	}
	db := utils.Db.Table(pojo.HostInfoTableName)
	db.Count(&result.Total)
	db.Limit(hostInfoSearch.PageSize).Offset(hostInfoSearch.CurrentPage * hostInfoSearch.PageSize).Find(&result.List)
	result.PageSize = hostInfoSearch.PageSize
	result.CurrentPage = hostInfoSearch.CurrentPage
	return result, err
}

func SetHostInfo(hostInfo pojo.HostInfo, hostInfoSet pojo.HostInfoSet) (result pojo.HostInfoBack, err error) {
	if hostInfo.HostName != utils.CsConfig.DefaultHost.HostName {
		return result, errors.New("域名错误")
	}
	var dbHostInfo pojo.HostInfo
	if hostInfoSet.ID == 0 {
		utils.Db.Where("host_name = ?", hostInfoSet.HostName).
			First(&dbHostInfo)
		if dbHostInfo.ID != 0 {
			return result, errors.New("域名不唯一")
		}
		_ = copier.Copy(&dbHostInfo, &hostInfoSet)
		dbHostInfo.AccessSecret = utils.CsConfig.DefaultHost.AccessSecret
		dbHostInfo.Salt = utils.CsConfig.DefaultHost.Salt
		dbHostInfo.PriKey = utils.CsConfig.DefaultHost.PriKey
		dbHostInfo.AccessExpire = utils.CsConfig.DefaultHost.AccessExpire
		err = utils.Db.Create(&dbHostInfo).Error
		if err != nil {
			return result, err
		}
		_, err = utils.InitTables(dbHostInfo.TablePrefix)
		err = utils.InitConfig(dbHostInfo)
		if err != nil {
			return result, err
		}
		err = utils.InitMenus(dbHostInfo)
	} else {
		utils.Db.Where("id = ?", hostInfoSet.ID).
			First(&dbHostInfo)
		if dbHostInfo.ID == 0 {
			return result, errors.New("更新的配置不存在")
		}
		hostInfoSet.HostName = ""
		hostInfoSet.TablePrefix = ""
		_ = copier.Copy(&dbHostInfo, hostInfoSet)
		dbHostInfo.Enabled = hostInfoSet.Enabled
		//dbHostInfoStr, _ := json.Marshal(dbHostInfo)
		//log.Printf("set dbHostInfo=%s", string(dbHostInfoStr))
		err = utils.Db.Save(&dbHostInfo).Error
	}
	utils.FlushTempHostInfo()
	return result, err
}

func DelHostInfo(id string) (result string, err error) {
	var dbHostInfo pojo.HostInfo
	utils.Db.Where("id = ?", id).Find(&dbHostInfo)
	if dbHostInfo.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	if dbHostInfo.HostName == utils.CsConfig.DefaultHost.HostName {
		return result, errors.New("基础数据不可删除")
	}
	utils.Db.Model(&dbHostInfo).Update("enabled", false)
	return "success", nil
}
