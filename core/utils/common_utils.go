package utils

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pojo"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const KeyDbPrefix = "prefix"             // db前缀
const KeyUserAwardCheck = "pu_uac_%d_%s" // redis 用户钱校验
const KeyLockUserAward = "pu_ua_%d"      // redis 用户钱变动锁
const KeyLockRequest = "pu_req_%d_%s"    // redis 请求锁
const KeyHostInfoTemp = "pu_hit_%s"      // redis 域名缓存key
const KeyConfigTemp = "pu_ct_%s_%s"      // redis 配置缓存key
const KeyUserCodeTemp = "pu_ut_%s_%s"    // redis 用户缓存key
const KeyUserTemp = "pu_ut_%s_%d"        // redis 用户缓存key
const KeySingle = "pu_so_"               // redis单点在线用户key前缀
const KeyRdOnline = "pu_online"          // redis在线用户key前缀
const KeyMqUserUpdate = "pu_uu"          // mqKey 用户信息更新通知
const KeyMqDeviceInfo = "pu_device_info" // mqKey 设备信息存储通知

func InWhiteIps(ip string) (result bool) {
	for _, whiteIp := range CsConfig.AwardIps {
		if whiteIp == ip {
			return true
		}
	}
	return false
}

func EncErrorBack(ctx *gin.Context, msg string) {
	message := I18nUtil.Translate(ctx, msg, nil)
	data := pojo.EncBackData{Message: message}
	dataStr, _ := json.Marshal(data)
	result, _ := DesEncrypt("c5ede42c", string(dataStr))
	ctx.JSON(http.StatusOK, pojo.EncResponse{
		Data: result,
	})
}

func EncSuccessBack(ctx *gin.Context, response string) {
	data := pojo.EncBackData{Message: response}
	dataStr, _ := json.Marshal(data)
	result, _ := DesEncrypt("c5ede42c", string(dataStr))
	ctx.JSON(http.StatusOK, pojo.EncResponse{
		Data: result,
	})
}

func IsValidAndroidId(id string) bool {
	androidIdReg := regexp.MustCompile(`^[a-f0-9]{16}$`)
	return androidIdReg.MatchString(id)
}

func ErrorBack(ctx *gin.Context, msg string) {
	msg = I18nUtil.Translate(ctx, msg, nil)
	ctx.JSON(http.StatusBadRequest, pojo.BaseResponse{
		Message: msg,
		Code:    500,
		Success: false,
	})
}

func ErrorObjBack(ctx *gin.Context, data interface{}, msg string) {
	msg = I18nUtil.Translate(ctx, msg, nil)
	ctx.JSON(http.StatusBadRequest, pojo.BaseResponse{
		Message: msg,
		Data:    data,
		Code:    500,
		Success: false,
	})
}

func ErrorMsgBack(ctx *gin.Context, msg string) {
	msg = I18nUtil.Translate(ctx, msg, nil)
	ctx.JSON(http.StatusOK, pojo.BaseResponse{
		Message: msg,
		Code:    500,
		Success: false,
	})
}

func SuccessBack(ctx *gin.Context, msg string) {
	msg = I18nUtil.Translate(ctx, msg, nil)
	ctx.JSON(http.StatusOK, pojo.BaseResponse{
		Message: msg,
		Code:    200,
		Success: true,
	})
}

func SuccessObjBack(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, pojo.BaseResponse{
		Message: "success",
		Data:    data,
		Code:    200,
		Success: true,
	})
}

func UnauthorizedBack(ctx *gin.Context, msg string) {
	msg = I18nUtil.Translate(ctx, msg, nil)
	ctx.JSON(http.StatusUnauthorized, pojo.BaseResponse{
		Message: msg,
		Code:    400,
		Success: false,
	})
}

func EncodePass(salt string, password string) string {
	saltedPass := password + salt
	passStr, _ := bcrypt.GenerateFromPassword([]byte(saltedPass), bcrypt.DefaultCost)
	return string(passStr)
}

func CheckPasswordHash(password string, hash string, salt string) bool {
	saltedPassword := salt + password
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}

func GetCurrentUser(ctx *gin.Context) (currentUser pojo.SysUser, err error) {
	userId := ctx.MustGet("userId").(int64)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	if userId == 0 {
		return currentUser, errors.New("token_error") // 登录已过期
	}
	currentUser = GetTempUser(hostInfo.TablePrefix, userId)
	//log.Printf("get CurrentUser=%v", userId)
	if currentUser.ID < 1 {
		return currentUser, errors.New("token_error") // 登录已过期
	}
	return currentUser, nil
}

func Test() {
}

func ContainsLink(input string) bool {
	regex := `https?://[^\s]+`
	re := regexp.MustCompile(regex)
	return re.MatchString(input)
}

func IsLocalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	localRanges := []net.IPNet{
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},    // 127.0.0.0/8
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},     // 10.0.0.0/8
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},  // 172.16.0.0/12
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)}, // 192.168.0.0/16
	}
	for _, r := range localRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

func GetIpInfoVip(ip string) (ipInfo string) {
	requestUrl := fmt.Sprintf("https://ipwhois.pro/json/%s?key=Yju4opPKnUUTOLlm", ip)
	data, _, err := ProxyGetRequestAll(requestUrl, nil, nil)
	if err != nil {
		return ipInfo
	}
	return string(data)
}

func IsValidPassword(str string) bool {
	pattern := `^[0-9A-Za-z!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]{6,20}$`
	match, _ := regexp.MatchString(pattern, str)
	return match
}

func IsPhone(str string) bool {
	pattern := `^\d{8,14}$`
	match, _ := regexp.MatchString(pattern, str)
	return match
}

func IsEmail(str string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, str)
	return match
}

func GetRegisterIPAddress(ctx *gin.Context) (result []string) {
	ip := ctx.GetHeader("x-forwarded-for")
	if ip == "" || ip == "unknown" {
		ip = ctx.GetHeader("Proxy-Client-IP")
	}
	if ip == "" || ip == "unknown" {
		ip = ctx.GetHeader("WL-Proxy-Client-IP")
	}
	if strings.Contains(ip, ",") {
		return strings.Split(ip, ",")
	}
	if ip == "" || ip == "unknown" {
		ip = ctx.ClientIP()
	}
	tempResults := strings.Split(ip, ",")
	result = make([]string, 0)
	for _, tempResult := range tempResults {
		result = append(result, strings.TrimSpace(tempResult))
	}
	return result
}

func GetIPAddress(ctx *gin.Context) string {
	ip := ctx.GetHeader("x-forwarded-for")
	if ip == "" || ip == "unknown" {
		ip = ctx.GetHeader("Proxy-Client-IP")
	}
	if ip == "" || ip == "unknown" {
		ip = ctx.GetHeader("WL-Proxy-Client-IP")
	}
	//fmt.Printf("ip1=%v\n", ip)
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ",")[0]
	}
	if ip == "" || ip == "unknown" {
		ip = ctx.ClientIP()
	}
	//fmt.Printf("ip2=%v\n", ip)
	return ip
}

func GetRequestHost(ctx *gin.Context) (host string) {
	return strings.Split(ctx.Request.Host, ":")[0]
}

func GetRequestFullHost(ctx *gin.Context) (host string) {
	scheme := "http"
	if ctx.Request.TLS != nil || ctx.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)
}

var GlobalConfig base.Config
var CsConfig base.CsConfig

func InStrings(datas []string, str string) bool {
	for _, v := range datas {
		if v == str {
			return true
		}
	}
	return false
}

func InInt64s(arr []int64, item int64) bool {
	for _, i := range arr {
		if i == item {
			return true
		}
	}
	return false
}

func CheckAppVersion(appVersion string, version int64) bool {
	reg := regexp.MustCompile(`\D`)
	appVersion = reg.ReplaceAllString(appVersion, "")
	floatVersion, _ := strconv.ParseInt(appVersion, 10, 64)
	//log.Printf("check app version %s->%d->%d", appVersion, floatVersion, version)
	return floatVersion >= version
}
