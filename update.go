package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var (
	jsonSlice = JsonInfoSlice{}
)

type JsonInfoSlice []*JsonUnit

func (s JsonInfoSlice) Len() int {
	return len(s)
}

func (s JsonInfoSlice) Less(i, j int) bool {
	m1 := s[i]
	m2 := s[j]

	v1, _ := strconv.Atoi(m1.Version)
	v2, _ := strconv.Atoi(m2.Version)

	if v1 < v2 {
		return true
	} else if v1 > v2 {
		return false
	} else {
		return false
	}
}

func (s JsonInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func getUpdateSqlFmt() (fmtStr string) {
	fmtStr = ""
	fmtStr += "\n"
	fmtStr += "USE `%v`;\n"
	fmtStr += "\n"
	fmtStr += "DELIMITER ;;\n"
	fmtStr += "CREATE PROCEDURE db_update()\n"
	fmtStr += "BEGIN\n"
	fmtStr += "    SELECT @id:=MAX(id) from db_version;\n"
	fmtStr += "    SELECT @db_ver:=version from db_version where id = @id;\n"
	fmtStr += "    \n"
	fmtStr += "    %v\n"
	fmtStr += "    \n"
	fmtStr += "    INSERT INTO db_version(version, update_time) VALUES(%v, utc_timestamp());\n"
	fmtStr += "END;;\n"
	fmtStr += "\n"
	fmtStr += "DELIMITER ;\n"
	fmtStr += "\n"
	fmtStr += "CALL db_update();\n"
	fmtStr += "\n"
	fmtStr += "DROP PROCEDURE db_update;\n"

	return
}

func loadJsonFile(path string, f os.FileInfo, err error) error {
	fmt.Println(fmt.Sprintf("loadJsonFile: %v", path))

	if f == nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
		return err
	}
	if f.IsDir() {
		return nil
	}
	ok := strings.HasSuffix(f.Name(), ".json")
	if !ok {
		return nil
	}

	ret, jsonInfo := initJsonForUpdate(path)
	if false == ret || nil == jsonInfo {
		err = fmt.Errorf("loadJsonFile, initJson failed")
		return err
	}

	fileVersionPostfix := getDbFileVersionPostfix(path)
	if fileVersionPostfix != jsonInfo.Version {
		err = fmt.Errorf("Error: fileVersionPostfix cant match to version in json")
		return err
	}

	if nil != jsonInfo {
		jsonSlice = append(jsonSlice, jsonInfo)
	}

	return nil
}

func ergodicFilelist(path string) {
	jsonSlice = JsonInfoSlice{}
	err := filepath.Walk(path, loadJsonFile)
	if err != nil {
		fmt.Printf("Error: filepath.Walk() returned %v\n", err)
		return
	}

	sort.Sort(jsonSlice)
}

// 生成更新sql文件============================================================================begin
func generateUpdateSql(path string) {
	ergodicFilelist(path)

	jsonSliceLen := len(jsonSlice)
	if jsonSliceLen <= 1 {
		fmt.Println("Error: json file is not more than one")
		return
	}

	dbName := jsonSlice[0].DbName
	maxVersion := jsonSlice[jsonSliceLen-1].Version

	updateSqlStr := ""

	for i := 0; i < jsonSliceLen-1; i++ {
		upSqlStr := generateTableUpdate(i, i+1)

		dbVer := jsonSlice[i].Version
		updateSqlStr += fmt.Sprintf("IF(@db_ver <= %v)THEN\n%v\nEND IF;\n", dbVer, upSqlStr)
	}

	UPDATE_SQL := getUpdateSqlFmt()
	totalSqlStr := fmt.Sprintf(UPDATE_SQL, dbName, updateSqlStr, maxVersion)

	writeSql(totalSqlStr, true)
}

func generateTableUpdate(i, j int) (upSqlStr string) {
	fmt.Println(fmt.Sprintf("generateTableUpdate, oldVer:%v, newVer:%v", i, j))

	upSqlStr = ""

	oldDbJson := jsonSlice[i]
	newDbJson := jsonSlice[j]

	for k, _ := range oldDbJson.Tables {
		isExistInNewDB := false
		for chkK, _ := range newDbJson.Tables {
			if k == chkK {
				isExistInNewDB = true
				break
			}
		}

		if false == isExistInNewDB {
			// 旧库中有而新库没有的表DROP掉
			upSqlStr += fmt.Sprintf("DROP TABLE %v;\n", k)
		}
	}
	for k, v := range newDbJson.Tables {
		isExistInOldDB := false
		for chkK, _ := range oldDbJson.Tables {
			if k == chkK {
				isExistInOldDB = true
				break
			}
		}

		if false == isExistInOldDB {
			// 旧库中没有而新库中有的表需要创建
			upSqlStr += generateTable(k, v)
		}
	}
	for k, oldTab := range oldDbJson.Tables {
		for chkK, newTab := range newDbJson.Tables {
			if k == chkK {
				// 旧库和新库中都有的表需要更新
				upSqlStr += generateFieldUpdate(oldTab, newTab, k)
			}
		}
	}

	upSqlStr += "\n"

	return
}

func generateFieldUpdate(oldTab TableUnit, newTab TableUnit, tabName string) (fieldUpSqlStr string) {
	fieldUpSqlStr = ""

	for _, oldField := range oldTab.Fields {
		oldFieldName := oldField.Name

		isExistInNewTab := false
		for _, newField := range newTab.Fields {
			newFieldName := newField.Name
			if newFieldName == oldFieldName {
				isExistInNewTab = true
				break
			}
		}

		if false == isExistInNewTab {
			// 老表中的字段在新表中没有
			fieldUpSqlStr += fmt.Sprintf("ALTER TABLE %v DROP %v;\n", tabName, oldFieldName)
		}
	}

	for _, newField := range newTab.Fields {
		newFieldName := newField.Name

		isExistInOldTab := false
		for _, oldField := range oldTab.Fields {
			oldFieldName := oldField.Name
			if newFieldName == oldFieldName {
				isExistInOldTab = true
				break
			}
		}

		if false == isExistInOldTab {
			// 新表中的字段在老表中没有
			addSqlStr, _ := generateField(newField, true)
			fieldUpSqlStr += fmt.Sprintf("ALTER TABLE %v ADD %v", tabName, addSqlStr)
		}
	}

	for _, oldField := range oldTab.Fields {
		oldFieldName := oldField.Name

		for _, newField := range newTab.Fields {
			newFieldName := newField.Name
			if newFieldName == oldFieldName {
				// 老表和新表中都有的字段
				isFieldSame := checkTabFieldIsSame(oldField, newField)
				if false == isFieldSame {
					modifySqlStr, _ := generateField(newField, true)
					fieldUpSqlStr += fmt.Sprintf("ALTER TABLE %v MODIFY %v", tabName, modifySqlStr)
				}
			}
		}
	}

	if fieldUpSqlStr != "" {
		fieldUpSqlStr += "\n"
	}

	return
}

func checkTabFieldIsSame(field1 FieldUnit, field2 FieldUnit) bool {
	if field1.Type != field2.Type {
		return false
	}

	if field1.Length != field2.Length {
		return false
	}

	if field1.AllowNull != field2.AllowNull {
		return false
	}

	if field1.Annotation != field2.Annotation {
		return false
	}

	if field1.AutoIncrement != field2.AutoIncrement {
		return false
	}

	if field1.DefaultVal != field2.DefaultVal {
		return false
	}

	if field1.Key != field2.Key {
		return false
	}

	if field1.Point != field2.Point {
		return false
	}

	return true
}

// 生成更新sql文件============================================================================end
