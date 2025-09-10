package utils

import (
	"BaseGoUni/core/pojo"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand/v2"
	"strconv"
	"time"
)

func GetRandomRangeSecond(min int, max int) (result time.Duration) {
	return time.Duration(min+rand.IntN(max-min)) * time.Second
}

func GetInt64Cache(preFix string, key string, defaultValue int64) (result int64) {
	redisKey := fmt.Sprintf(KeyConfigTemp, preFix, key)
	dataStr := ""
	data := RD.Get(context.Background(), redisKey)
	if data != nil && data.Err() == nil {
		dataStr = data.Val()
	}
	newData := false
	if dataStr == "" {
		db := NewPrefixDb(preFix)
		var sysConfig pojo.SysConfig
		db.Where("config_key = ?", key).First(&sysConfig)
		if sysConfig.ID != 0 {
			dataStr = sysConfig.ConfigValue
			newData = true
		}
	}
	if dataStr != "" {
		num, err := strconv.ParseInt(dataStr, 10, 64)
		if err == nil {
			return num
		}
	}
	if newData {
		FlushInt64Cache(preFix, key, defaultValue)
	}
	return defaultValue
}

func FlushInt64Cache(preFix string, key string, value int64) (result int64) {
	redisKey := fmt.Sprintf(KeyConfigTemp, preFix, key)
	db := NewPrefixDb(preFix)
	var sysConfig pojo.SysConfig
	db.Where("config_key = ?", key).First(&sysConfig)
	if sysConfig.ID == 0 {
		sysConfig.ConfigKey = key
		sysConfig.ConfigValue = strconv.FormatInt(value, 10)
		db.Create(&sysConfig)
		RD.SetEX(context.Background(), redisKey, value, GetRandomRangeSecond(20*60, 40*60))
		return value
	}
	db.Model(&sysConfig).Update("config_value", value)
	RD.SetEX(context.Background(), redisKey, value, GetRandomRangeSecond(20*60, 40*60))
	return value
}

func GetStringCache(preFix string, key string, defaultValue *string) (result *string) {
	redisKey := fmt.Sprintf(KeyConfigTemp, preFix, key)
	dataStr := ""
	data := RD.Get(context.Background(), redisKey)
	if data != nil && data.Err() == nil {
		dataStr = data.Val()
		return &dataStr
	}
	db := NewPrefixDb(preFix)
	var sysConfig pojo.SysConfig
	db.Where("config_key = ?", key).First(&sysConfig)
	if sysConfig.ID != 0 {
		return &sysConfig.ConfigValue
	}
	FlushStringCache(preFix, key, *defaultValue)
	return defaultValue
}

func FlushStringCache(preFix string, key string, value string) (result *string) {
	redisKey := fmt.Sprintf(KeyConfigTemp, preFix, key)
	db := NewPrefixDb(preFix)
	var sysConfig pojo.SysConfig
	db.Where("config_key = ?", key).First(&sysConfig)
	if sysConfig.ID == 0 {
		sysConfig.ConfigKey = key
		sysConfig.ConfigValue = value
		db.Create(&sysConfig)
		RD.SetEX(context.Background(), redisKey, value, GetRandomRangeSecond(20*60, 40*60))
		return &value
	}
	db.Model(&sysConfig).Update("config_value", value)
	result = &value
	RD.SetEX(context.Background(), redisKey, value, GetRandomRangeSecond(20*60, 40*60))
	return result
}

func GetUserUniKey(preFix string) (result string) {
	uniKey := RandomString(4)
	tempUser := GetTempUserCode(preFix, uniKey)
	for i := 0; tempUser.ID != 0 && i < 30; i++ {
		uniKey = RandomString(4)
		tempUser = GetTempUserCode(preFix, uniKey)
	}
	if tempUser.ID == 0 {
		return uniKey
	}
	return result
}

func GetOrLoad[T any](
	ctx context.Context,
	db *gorm.DB,
	redisKey string,
	expire time.Duration,
	loadFromDB func(*gorm.DB) (T, error),
) (T, error) {
	var result T
	data, err := RD.Get(ctx, redisKey).Result()
	if err == nil && data != "" {
		if e := json.Unmarshal([]byte(data), &result); e == nil {
			return result, nil
		}
	}
	result, err = loadFromDB(db)
	if err != nil {
		return result, err
	}
	if jsonStr, e := json.Marshal(result); e == nil {
		_ = RD.SetEX(ctx, redisKey, jsonStr, expire).Err()
	}
	return result, nil
}

//func GetTempUsers(preFix string, userId int64) (result []pojo.SysUser) {
//	redisKey := fmt.Sprintf(KeyUserTemp, preFix, userId)
//	result, _ = GetOrLoad(context.Background(), Db, redisKey, GetRandomRangeSecond(20*60, 40*60),
//		func(db *gorm.DB) (tempResult []pojo.SysUser, err error) {
//			if err = db.Where("user_id = ?", userId).Find(&tempResult).Error; err != nil {
//				return tempResult, err
//			}
//			if len(tempResult) == 0 {
//				err = errors.New("no match data")
//			}
//			return tempResult, err
//		},
//	)
//	return result
//}

func GetTempUserCode(preFix string, uniKey string) (result pojo.SysUser) {
	redisKey := fmt.Sprintf(KeyUserCodeTemp, preFix, uniKey)
	result, _ = GetOrLoad(context.Background(), NewPrefixDb(preFix), redisKey, GetRandomRangeSecond(20*60, 40*60),
		func(db *gorm.DB) (tempResult pojo.SysUser, err error) {
			if err = db.Where("uni_key = ?", uniKey).First(&tempResult).Error; err != nil {
				return tempResult, err
			}
			if tempResult.ID == 0 {
				err = errors.New("no match data")
			}
			return tempResult, err
		},
	)
	return result
}

func GetTempUser(preFix string, userId int64) (result pojo.SysUser) {
	redisKey := fmt.Sprintf(KeyUserTemp, preFix, userId)
	result, _ = GetOrLoad(context.Background(), NewPrefixDb(preFix), redisKey, GetRandomRangeSecond(20*60, 40*60),
		func(db *gorm.DB) (tempResult pojo.SysUser, err error) {
			if err = db.Where("id = ?", userId).First(&tempResult).Error; err != nil {
				return tempResult, err
			}
			if tempResult.ID == 0 {
				err = errors.New("no match data")
			}
			return tempResult, err
		},
	)
	return result
}

func UpdateTempUser(preFix string, user pojo.SysUser) {
	redisKey := fmt.Sprintf(KeyUserTemp, preFix, user.ID)
	jsonStr, _ := json.Marshal(user)
	RD.SetEX(context.Background(), redisKey, string(jsonStr), GetRandomRangeSecond(20*60, 40*60))
	redisKey = fmt.Sprintf(KeyUserCodeTemp, preFix, user.UniKey)
	RD.SetEX(context.Background(), redisKey, string(jsonStr), GetRandomRangeSecond(20*60, 40*60))
}

func FlushTempUser(preFix string, userId int64) {
	var user pojo.SysUser
	db := NewPrefixDb(preFix)
	db.Where("id = ?", userId).First(&user)
	if user.ID == 0 {
		return
	}
	UpdateTempUser(preFix, user)
}

func GetTempHostInfo(hostName string) (result pojo.HostInfo) {
	redisKey := fmt.Sprintf(KeyHostInfoTemp, hostName)
	result, _ = GetOrLoad(context.Background(), Db, redisKey, GetRandomRangeSecond(20*60, 40*60),
		func(db *gorm.DB) (tempResult pojo.HostInfo, err error) {
			if err = Db.Where("host_name = ?", hostName).First(&tempResult).Error; err != nil {
				return tempResult, err
			}
			if tempResult.ID == 0 {
				err = errors.New("no match data")
			}
			return tempResult, err
		},
	)
	return result
}

func FlushTempHostInfo() {
	var hostInfos []pojo.HostInfo
	Db.Find(&hostInfos)
	for _, hostInfo := range hostInfos {
		redisKey := fmt.Sprintf(KeyHostInfoTemp, hostInfo.HostName)
		jsonStr, _ := json.Marshal(hostInfo)
		RD.SetEX(context.Background(), redisKey, string(jsonStr), GetRandomRangeSecond(20*60, 40*60))
	}
}
