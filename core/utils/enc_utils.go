package utils

import (
	"BaseGoUni/core/base"
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jinzhu/copier"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var encKey = "enc_key_nj"

func DesEncrypt(encKey, source string) (string, error) {
	block, err := getDesCipher(encKey, source)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, []byte(encKey))
	paddedSource := pkcs5Padding([]byte(source), block.BlockSize())
	ciphertext := make([]byte, len(paddedSource))
	mode.CryptBlocks(ciphertext, paddedSource)
	return hex.EncodeToString(ciphertext), nil
}

func CheckEncReq(encRequest base.EncData, ip string) (result base.DeviceInfo, err error) {
	realKey := strings.ToLower(GetMd5(fmt.Sprintf("%s_rg_%d", encRequest.EncData, encRequest.Time)))
	if encRequest.CheckKey != realKey {
		log.Printf("checkKey=%s;realKey=%s", encRequest.CheckKey, realKey)
		return result, errors.New("checkKey error")
	}
	encodeKey := strings.ToLower(GetMd58(encKey + strconv.FormatInt(encRequest.Time, 10)))
	realStr, err := desDecrypt(encodeKey, encRequest.EncData)
	if err != nil {
		return result, err
	}
	_ = json.Unmarshal([]byte(realStr), &result)
	now := time.Now()
	requestTime := time.UnixMilli(encRequest.Time)
	if now.Sub(requestTime) > 24*time.Hour {
		log.Printf("加密请求时间戳错误:%d;%s", encRequest.Time, result.Data)
	}
	var tempDevice base.DeviceInfo
	_ = copier.Copy(&tempDevice, result)
	tempDevice.Data = ""
	tempDeviceStr, _ := json.Marshal(tempDevice)
	go func() {
		_ = PublishMQ(MQMessage{
			MessageType: KeyMqDeviceInfo,
			Data:        string(tempDeviceStr),
			DataMore:    ip,
		})
	}()
	return result, nil
}

func pkcs5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

func getDesCipher(encKey, source string) (cipher.Block, error) {
	if source == "" {
		return nil, fmt.Errorf("source cannot be empty")
	}
	block, err := des.NewCipher([]byte(encKey))
	if err != nil {
		return nil, err
	}
	return block, nil
}

func pkcs5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func desDecrypt(encKey, source string) (string, error) {
	block, err := getDesCipher(encKey, source)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, []byte(encKey))
	ciphertext, err := hex.DecodeString(source)
	if err != nil {
		return "", err
	}
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	plaintext = pkcs5UnPadding(plaintext)
	return string(plaintext), nil
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func DecPriKey(encryptedStr string, privateKeyData string) (resp string, err error) {
	//log.Printf("encryptedStr:%s\n privateKeyData:%s", encryptedStr, privateKeyData)
	priBlock, _ := pem.Decode([]byte(privateKeyData))
	privateKey, priErr := x509.ParsePKCS1PrivateKey(priBlock.Bytes)
	if priErr != nil {
		log.Printf("Load private key error privateKeyData=%s", privateKeyData)
		return "", priErr
	}
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		fmt.Println("Failed to base64 decode encrypted string:", err)
		return "", err
	}
	decryptedData, err := rsa.DecryptPKCS1v15(nil, privateKey, encryptedData)
	if err != nil {
		fmt.Println("Failed to decrypt data:", err)
		return "", err
	}
	return string(decryptedData), nil
}

func GetJwtToken(accessSecret string, accessExpire int64, username string, userId int64, userType int, hostName string) (string, error) {
	iat := time.Now().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + accessExpire
	claims["iat"] = iat
	claims["username"] = username
	claims["userId"] = userId
	claims["userType"] = userType
	claims["hostName"] = hostName
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(accessSecret))
}

func GetMerchantJwtToken(accessSecret string, accessExpire int64, username string, userId int64, userType int, hostName string, childCode string) (string, error) {
	iat := time.Now().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + accessExpire
	claims["iat"] = iat
	claims["username"] = username
	claims["userId"] = userId
	claims["userType"] = userType
	claims["hostName"] = hostName
	claims["childCode"] = childCode
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(accessSecret))
}

func ParseToken(accessSecret string, tokenString string) (userId int64, userType int, hostName string, childCode string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 返回与创建 token 时相同的密钥
		return []byte(accessSecret), nil
	})
	if err != nil {
		return userId, userType, hostName, childCode, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIdValue, ok := claims["userId"]
		if !ok {
			return userId, userType, hostName, childCode, fmt.Errorf("token does not contain valid userId")
		}
		userTypeValue, ok := claims["userType"]
		if !ok {
			return userId, userType, hostName, childCode, fmt.Errorf("token does not contain valid userType")
		}
		hostNameValue, ok := claims["hostName"]
		if !ok {
			return userId, userType, hostName, childCode, fmt.Errorf("token does not contain valid hostName")
		}
		var realId int64
		switch userIdObj := userIdValue.(type) {
		case float64:
			realId = int64(userIdObj)
		default:
			return userId, userType, hostName, childCode, fmt.Errorf("userId is of a type I don't understand")
		}
		var realType int
		switch userTypeObj := userTypeValue.(type) {
		case float64:
			realType = int(userTypeObj)
		default:
			return userId, userType, hostName, childCode, fmt.Errorf("userType is of a type I don't understand")
		}
		switch hostNameObj := hostNameValue.(type) {
		case string:
			hostName = hostNameObj
		default:
			return userId, userType, hostName, childCode, fmt.Errorf("userId is of a type I don't understand")
		}
		childCodeValue, _ := claims["childCode"]
		switch childCodeObj := childCodeValue.(type) {
		case string:
			childCode = childCodeObj
		}
		return realId, realType, hostName, childCode, nil
	} else {
		return userId, userType, hostName, childCode, fmt.Errorf("invalid token")
	}
}

func GetMd58(data string) string {
	return getMd5(data, 8)
}
func GetMd516(data string) string {
	return getMd5(data, 16)
}

func GetMd5(data string) string {
	return getMd5(data, 32)
}

func getMd5(data string, len int) string {
	hash := md5.New()
	hash.Write([]byte(data))
	md5Bytes := hash.Sum(nil)
	md5String := hex.EncodeToString(md5Bytes)
	return md5String[:len]
}

func EncCheck(securityKey string, sign string, reqData any) (result bool) {
	realSign, paramsStr := GetSign(reqData, securityKey)
	result = realSign == sign
	if !result {
		log.Printf("sign=%s;realSign=%s;params=%s", sign, realSign, string(paramsStr))
	}
	return result
}

func GetSign(data any, key string) (result string, paramsStr []byte) {
	params := make(map[string]string)
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)
		if fieldValue.String() == "" {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "sign" {
			continue
		}
		var fieldStr string
		switch fieldValue.Kind() {
		case reflect.String:
			fieldStr = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldStr = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldStr = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			fieldStr = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.Bool:
			fieldStr = strconv.FormatBool(fieldValue.Bool())
		default:
			continue
		}
		params[jsonTag] = fieldStr
	}
	paramsStr, _ = json.Marshal(params)
	return GetMapSign(params, key), paramsStr
}

func GetMapSign(params map[string]string, key string) string {
	var keys []string
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var data []string
	for _, k := range keys {
		data = append(data, fmt.Sprintf("%s=%s", k, params[k]))
	}
	dataString := strings.Join(data, "&")
	dataString = dataString + key
	//log.Printf("\n\n\ndataString=%s;\n\n\n", dataString)
	return strings.ToLower(MD5(dataString))
}
