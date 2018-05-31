package main

import (
	"fmt"
	"strconv"
)

var (
	CREATE_DATABASE    = "DROP DATABASE IF EXISTS `%v`;\nCREATE DATABASE `%v` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;\n"
	ADD_PRIVILEGES     = "GRANT ALL PRIVILEGES ON %v.* TO '%v'@'localhost';\nFLUSH PRIVILEGES;\n"
	USE_DB             = "USE `%v`;\n"
	CREATE_TABLE_BEGIN = "CREATE TABLE IF NOT EXISTS `%v` (\n"
	CREATE_TABLE_END   = "PRIMARY KEY (%v)\n) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='%v';"
	FIELD              = "`%v` %v%v %v %v %v %v COMMENT '%v',\n"
	FIELD_FOR_UPDATE   = "`%v` %v%v %v %v %v %v COMMENT '%v';\n"
)

// 生成创建sql文件============================================================================begin
func generateCreateSql(fileName string) bool {
	ret := initJson(fileName)
	if false == ret {
		return false
	}

	fileVersionPostfix := getDbFileVersionPostfix(fileName)
	if fileVersionPostfix != jsonCfg.Version {
		fmt.Println("Error: fileVersionPostfix cant match to version in json")
		return false
	}

	sqlStr := ""

	// create DATABASE
	sqlStr += fmt.Sprintf(CREATE_DATABASE, jsonCfg.DbName, jsonCfg.DbName)
	// add privileges
	sqlStr += fmt.Sprintf(ADD_PRIVILEGES, jsonCfg.DbName, jsonCfg.DbUser)
	sqlStr += fmt.Sprintf(USE_DB, jsonCfg.DbName)
	sqlStr += "\n"

	// create table
	for tabName, tabInfo := range jsonCfg.Tables {
		sqlStr += generateTable(tabName, tabInfo)
	}
	sqlStr += "\n"

	// add version info
	sqlStr += "INSERT INTO db_version(version, update_time) VALUES(" + jsonCfg.Version + ", utc_timestamp());"
	sqlStr += "\n"

	writeSql(sqlStr, false)

	return true
}

func generateTable(tabName string, tabInfo TableUnit) (tabSqlStr string) {
	fmt.Println("generateTable, tabName:", tabName)

	tabSqlStr = ""

	tabSlice := []string{}
	if tabInfo.Sharding > 1 {
		// 有分表
		for i := 1; i <= tabInfo.Sharding; i++ {
			tmpTabName := fmt.Sprintf("%v_%v", tabName, i)
			tabSlice = append(tabSlice, tmpTabName)
		}
	} else {
		tabSlice = append(tabSlice, tabName)
	}

	for _, tab := range tabSlice {
		tabSqlStr += fmt.Sprintf(CREATE_TABLE_BEGIN, tab)
		keys := []string{}
		for _, fieldInfo := range tabInfo.Fields {
			fieldSqlStr, keyList := generateField(fieldInfo, false)
			tabSqlStr += fieldSqlStr
			keys = append(keys, keyList...)
		}

		if len(keys) < 1 {
			fmt.Println("Error: cant find key in table:", tab)
			return
		}

		keyStr := ""
		keysLen := len(keys)
		for idx, key := range keys {
			keyStr += key
			if idx+1 < keysLen {
				keyStr += ","
			}
		}
		tabSqlStr += fmt.Sprintf(CREATE_TABLE_END, keyStr, tabInfo.Annotation)
		tabSqlStr += "\n\n"
	}

	return
}

func generateField(fieldInfo FieldUnit, isUpdate bool) (fieldSqlStr string, keys []string) {
	fmt.Println("generateField, fieldName:", fieldInfo.Name)

	fieldSqlStr = ""
	keys = []string{}

	if fieldInfo.Name == "" {
		fmt.Println("Error: fieldName is nil")
		return
	}

	// check type
	if false == checkIsLegalType(fieldInfo.Type) {
		fmt.Println(fmt.Sprintf("Error: fieldType:%v is invalid", fieldInfo.Type))
		return
	}

	// is key
	if fieldInfo.Key == "1" {
		keys = append(keys, fieldInfo.Name)
	}

	// length
	lengthSqlStr := ""
	if fieldInfo.Type == "string" || fieldInfo.Type == "bytearray" {
		if fieldInfo.Length == "" {
			fmt.Println(fmt.Sprintf("Error: fieldType:%v length:%v is invalid", fieldInfo.Type, fieldInfo.Length))
			return
		}
		fieldLen, err := strconv.Atoi(fieldInfo.Length)
		if nil != err || fieldLen <= 0 {
			fmt.Println(fmt.Sprintf("Error: fieldType:%v length:%v is invalid", fieldInfo.Type, fieldInfo.Length))
			return
		}

		lengthSqlStr = fmt.Sprintf("(%v)", fieldInfo.Length)
	}
	if fieldInfo.Type == "float" || fieldInfo.Type == "double" || fieldInfo.Type == "decimal" {
		fieldLen, err1 := strconv.Atoi(fieldInfo.Length)
		pointLen, err2 := strconv.Atoi(fieldInfo.Point)
		if nil == err1 && fieldLen > 0 && nil == err2 && pointLen > 0 {
			lengthSqlStr = fmt.Sprintf("(%v,%v)", fieldLen, pointLen)
		}
	}

	// allow_null
	allowNullSqlStr := ""
	if fieldInfo.AllowNull != "1" {
		allowNullSqlStr = "NOT NULL"
	}

	//type and unsigned
	typeName := ""
	unsigned := ""
	fieldType := fieldInfo.Type
	if fieldType == "int8" {
		typeName = "TINYINT"
	} else if fieldType == "uint8" {
		typeName = "TINYINT"
		unsigned = "UNSIGNED"
	} else if fieldType == "int16" {
		typeName = "SMALLINT"
	} else if fieldType == "uint16" {
		typeName = "SMALLINT"
		unsigned = "UNSIGNED"
	} else if fieldType == "int32" {
		typeName = "INT"
	} else if fieldType == "uint32" {
		typeName = "INT"
		unsigned = "UNSIGNED"
	} else if fieldType == "int64" {
		typeName = "BIGINT"
	} else if fieldType == "uint64" {
		typeName = "BIGINT"
		unsigned = "UNSIGNED"
	} else if fieldType == "float" {
		typeName = "FLOAT"
	} else if fieldType == "double" {
		typeName = "DOUBLE"
	} else if fieldType == "decimal" {
		typeName = "DECIMAL"
	} else if fieldType == "string" {
		typeName = "VARCHAR"
	} else if fieldType == "bytearray" {
		typeName = "BLOB"
	} else if fieldType == "timestamp" {
		typeName = "TIMESTAMP"
	} else {
		fmt.Println("Error: cant find the field type")
		return
	}

	// auto increment
	autoIncrementSqlStr := ""
	if fieldInfo.AutoIncrement == "1" {
		autoIncrementSqlStr = "AUTO_INCREMENT"
	}

	// default
	defaultSqlStr := ""
	if fieldInfo.DefaultVal != "" {
		defaultSqlStr = "DEFAULT " + "'" + fieldInfo.DefaultVal + "'"
	}

	// comment
	commentSqlStr := ""
	if fieldInfo.Annotation != "" {
		commentSqlStr = fieldInfo.Annotation
	}

	fieldSqlFmt := FIELD
	if true == isUpdate {
		fieldSqlFmt = FIELD_FOR_UPDATE
	}

	fieldSqlStr += fmt.Sprintf(fieldSqlFmt, fieldInfo.Name, typeName, lengthSqlStr, unsigned, allowNullSqlStr, autoIncrementSqlStr, defaultSqlStr, commentSqlStr)

	return
}

// 生成创建sql文件============================================================================end
