package utils

import (
	"BaseGoUni/core/pojo"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hash/fnv"
	"log"
	"reflect"
	"regexp"
	"strings"
)

type ShardRule struct {
	ShardField   string
	ShardCount   int
	TableName    string
	AllTableName string
}

var shardRules = map[reflect.Type]*ShardRule{
	reflect.TypeOf(pojo.CashHistory{}): {
		ShardField:   pojo.CashHistoryShardingName,
		ShardCount:   pojo.CashHistoryShards,
		TableName:    pojo.CashHistoryTableName,
		AllTableName: pojo.AllCashHistoryShardingName,
	},
}

func getFieldVale(value interface{}) (result interface{}) {
	indirectValue := reflect.Indirect(reflect.ValueOf(value))
	switch v := indirectValue.Interface().(type) {
	case pojo.CashHistory:
		return v.UserId
	default:
		log.Printf("not support type: %T;%v\n", value, v)
		return nil
	}
}

func getQueryVale(tx *gorm.DB, rule ShardRule) (result interface{}) {
	paramMap := make(map[string]interface{})
	where := tx.Statement.Clauses["WHERE"].Expression
	if where == nil {
		log.Printf("getQueryVale where nil")
		return nil
	}
	paramMap = extractParams(where.(clause.Where))
	for k, v := range paramMap {
		if k == rule.ShardField {
			return v
		}
	}
	log.Printf("getQueryVale no match query value.where=%v", where)
	return nil
}

func InitShardingHook(tx *gorm.DB) {
	tx.Callback().Create().Before("gorm:create").Register("custom_create_hook", BeforeCudHook)
	tx.Callback().Update().Before("gorm:update").Register("custom_update_hook", BeforeCudHook)
	tx.Callback().Delete().Before("gorm:delete").Register("custom_delete_hook", BeforeCudHook)
	tx.Callback().Query().Before("gorm:query").Register("custom_query_hook", BeforeRHook)
}

func GetShardingTableName(fieldValue interface{}) (result string) {
	rule, ok := shardRules[reflect.TypeOf(fieldValue)]
	if !ok {
		return result
	}
	return GetShardingTable(rule.TableName, getFieldVale(fieldValue), false)
}

func GetShardingTable(tableName string, fieldValue interface{}, isArray bool) string {
	if isArray {
		val := reflect.Indirect(reflect.ValueOf(fieldValue))
		tableNames := make([]string, 0)
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() != reflect.Struct {
				continue
			}
			tempValue := getFieldVale(elem.Interface())
			if tempValue == nil {
				//log.Printf("on table name=%s;elem=%v;type=%v", tableName, elem, reflect.TypeOf(elem))
				return tableName
			}
			tempName := GetShardingTable(tableName, tempValue, false)
			if len(tableNames) > 0 && !InStrings(tableNames, tempName) {
				log.Printf("no in same sharding table.tableName=%s;tableNames=%v;tempName=%s", tableName, tableNames, tempName)
				return tableName
			} else {
				tableNames = append(tableNames, tempName)
			}
		}
		if len(tableNames) == 0 {
			//log.Printf("on table name=%s;table names empty", tableName)
			return tableName
		}
		return tableNames[0]
	}
	rule := getRuleByTableName(tableName)
	if rule == nil {
		return tableName
	}
	var shardIndex int
	switch v := fieldValue.(type) {
	case int64:
		shardIndex = int(v % int64(rule.ShardCount))
	case string:
		shardIndex = int(hashString(v)) % rule.ShardCount
	default:
		return tableName
	}
	return GetShardingTableNameInt(tableName, int64(shardIndex), rule.ShardCount)
}

func BeforeCudHook(tx *gorm.DB) {
	rule := getRuleByTableName(tx.Statement.Table)
	if rule == nil {
		//log.Printf("on cud models table=%s;", tx.Statement.Table)
		return
	}
	var fieldValue interface{}
	model := tx.Statement.Model
	val := reflect.Indirect(tx.Statement.ReflectValue)
	isArray := val.Kind() == reflect.Slice || val.Kind() == reflect.Array
	if isArray {
		if val.Len() > 0 {
			_, ok := shardRules[reflect.TypeOf(val.Index(0).Interface())]
			if !ok {
				//log.Printf("on cud models table=%s;data[0]=%v is not supported;", rule.TableName,
				//	val.Index(0))
				tx.Table(rule.TableName)
				return
			}
			fieldValue = tx.Statement.Dest
		} else {
			//log.Println("on cud: empty array")
			tx.Table(rule.TableName)
			return
		}
	} else {
		fieldValue = getFieldVale(model)
		if fieldValue == nil {
			//log.Printf("on cud model=%v;", model)
			tx.Table(rule.TableName)
			return
		}
		if fieldValue == "" {
			fieldValue = getQueryVale(tx, *rule)
		}
	}
	tableName := GetShardingTable(rule.TableName, fieldValue, isArray)
	tx.Statement.Table = tableName
	tx.Table(tableName)
	//log.Printf("on cud value=%v;endTable=%s", fieldValue, tableName)
}

func BeforeRHook(tx *gorm.DB) {
	rule := getRuleByTableName(tx.Statement.Table)
	if rule == nil {
		//log.Printf("on r models table=%s;", tx.Statement.Table)
		return
	}
	paramMap := make(map[string]interface{})
	where := tx.Statement.Clauses["WHERE"].Expression
	if where == nil {
		tx.Statement.Table = rule.AllTableName
		//log.Printf("on r where nil.table=%s;", rule.AllTableName)
		tx.Table(rule.AllTableName)
		return
	}
	paramMap = extractParams(where.(clause.Where))
	var fieldValue interface{}
	haveKey := false
	for k, v := range paramMap {
		if k == rule.ShardField {
			fieldValue = v
			haveKey = true
			break
		}
	}
	var tableName string
	if haveKey {
		tableName = GetShardingTable(rule.TableName, fieldValue, false)
	} else {
		tableName = rule.AllTableName
	}
	//fieldValueStr, _ := json.Marshal(fieldValue)
	tx.Statement.Table = tableName
	tx.Table(tableName)
	//log.Printf("on r.tableName=%s;fieldValue=%s", tableName, string(fieldValueStr))
}

func getRuleByTableName(tableName string) (result *ShardRule) {
	for _, v := range shardRules {
		if v.TableName == tableName {
			return v
		}
	}
	return nil
}

func InitShardingDataBase(Db *gorm.DB, project interface{}, tableName string, shardingCount int) (err error) {
	for i := 0; i < shardingCount; i++ {
		realTableName := GetShardingTableNameInt(tableName, int64(i), shardingCount)
		err = Db.Table(realTableName).AutoMigrate(&project)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateView(shardCount uint, viewName string, tableSuffix string) {
	var unionParts []string
	var i uint
	for i = 0; i < shardCount; i++ {
		unionParts = append(unionParts, fmt.Sprintf("SELECT * FROM %s",
			GetShardingTableNameInt(tableSuffix, int64(i), int(shardCount))))
	}
	createViewSQL := fmt.Sprintf("CREATE VIEW %s AS %s;", viewName, strings.Join(unionParts, " UNION ALL "))
	_ = Db.Exec(createViewSQL)
}

func GetShardingTableNameInt(tableName string, index int64, shardingCount int) (result string) {
	if shardingCount > 10000 {
		panic(errors.New("sharding count too large"))
	} else if shardingCount > 1000 {
		result = fmt.Sprintf("%s_%04d", tableName, index%int64(shardingCount))
	} else if shardingCount > 100 {
		result = fmt.Sprintf("%s_%03d", tableName, index%int64(shardingCount))
	} else if shardingCount > 10 {
		result = fmt.Sprintf("%s_%02d", tableName, index%int64(shardingCount))
	} else {
		result = fmt.Sprintf("%s_%d", tableName, index%int64(shardingCount))
	}
	//log.Printf("GetShardingTableNameInt:tableName=%s;index=%d;shardingCount=%d;result=%s", tableName, index, shardingCount, result)
	return result
}

func extractParams(where clause.Where) map[string]interface{} {
	paramMap := make(map[string]interface{})
	re := regexp.MustCompile(`(\w+)\s*[=<>!]+\s*\?`)
	for _, cond := range where.Exprs {
		if eqCond, ok := cond.(clause.Expr); ok && eqCond.SQL != "" {
			matches := re.FindAllStringSubmatch(eqCond.SQL, -1)
			var keys []string
			for _, match := range matches {
				keys = append(keys, match[1])
			}
			if len(keys) <= len(eqCond.Vars) {
				for i, key := range keys {
					paramMap[key] = eqCond.Vars[i]
				}
			} else {
				log.Printf("params not match:keys=%v,vars=%v", keys, eqCond.Vars)
			}
		}
	}
	return paramMap
}

func hashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	result := h.Sum32()
	//log.Printf("hashString:%s=%d", s, result)
	return result
}
