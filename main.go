package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	defaultPath     = "./database"
	defaultJsonFile = "./database/db_v1.json"
	jsonCfg         = JsonUnit{}

	LegalTypeSlice = []string{"int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "float", "double", "decimal", "string", "bytearray", "timestamp"}
)

const (
	FileDir = "./sql/"
)

type JsonUnit struct {
	DbName  string               `json:"name"`    // 数据库名称
	Version string               `json:"version"` // 版本
	DbUser  string               `json:"definer"` // 数据库用户
	Tables  map[string]TableUnit `json:"table"`   // 数据库表列表
}

type TableUnit struct {
	Annotation string      `json:"comment"`          // 注释(描述)
	Ignore     string      `json:"ignore,omitempty"` // 是否忽略
	Sharding   int         `json:"sharding"`         // 分表数量(1代表不分)
	Fields     []FieldUnit `json:"fields"`           // 字段列表
}

type FieldUnit struct {
	Name          string `json:"name"`                     // 字段名
	Type          string `json:"type"`                     // 字段类型
	Length        string `json:"length,omitempty"`         // 字段长度
	AllowNull     string `json:"allow_null,omitempty"`     // 是否允许为空
	Point         string `json:"point,omitempty"`          // 小数点后几位
	Key           string `json:"key,omitempty"`            // 是否主键
	AutoIncrement string `json:"auto_increment,omitempty"` // 是否自增
	DefaultVal    string `json:"default,omitempty"`        // 默认值
	Annotation    string `json:"comment"`                  // 注释(描述)
}

func main() {
	generateDir()
	generateCreateSql(defaultJsonFile)
	generateUpdateSql(defaultPath)
}

// 文件操作相关==========================================================================begin
func checkPathIsExist(path string) bool {
	var exist = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func generateDir() {
	if checkPathIsExist(FileDir) {
		//如果文件夹存在,删除文件夹
		err := os.RemoveAll(FileDir)
		if err != nil {
			fmt.Println(fmt.Printf("Error: dir:%v, remove error:%s", FileDir, err))
			return
		}
	}

	os.Mkdir(FileDir, os.ModePerm)
}

func writeSql(sqlStr string, isUpdate bool) {
	// 生成文件路径和名称
	sqlFile := ""
	if isUpdate == true {
		sqlFile = FileDir + fmt.Sprintf("%v_update.sql", jsonCfg.DbName)
	} else {
		sqlFile = FileDir + fmt.Sprintf("%v_create.sql", jsonCfg.DbName)
	}

	// 创建文件
	if checkPathIsExist(sqlFile) == false {
		file, err := os.Create(sqlFile)
		if err != nil {
			fmt.Println(fmt.Printf("Error: file:%v, create error:%s", sqlFile, err))
			return
		}
		defer file.Close()
	}

	// 以只写的模式，打开文件
	f, err := os.OpenFile(sqlFile, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Printf("Error: file:%v, open error:%s", sqlFile, err))
		return
	}
	defer f.Close()

	// 查找文件末尾的偏移量
	n, _ := f.Seek(0, os.SEEK_END)
	// 从末尾的偏移量开始写入内容
	_, err = f.WriteAt([]byte(sqlStr), n)
	if err != nil {
		fmt.Println(fmt.Printf("Error: file:%v, write error:%s", sqlFile, err))
		return
	}

	fmt.Println(fmt.Sprintf("congratulation! file:%v, generate ok!", sqlFile))
}

func initJson(fileName string) bool {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
		return false
	}

	err = json.Unmarshal(bytes, &jsonCfg)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
		return false
	}

	//fmt.Println(jsonCfg)

	return true
}

func initJsonForUpdate(fileName string) (ret bool, jsonInfo *JsonUnit) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
		return false, nil
	}

	jsonUnitInfo := &JsonUnit{}
	err = json.Unmarshal(bytes, jsonUnitInfo)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
		return false, nil
	}

	//fmt.Println(*jsonUnitInfo)

	return true, jsonUnitInfo
}

// 文件操作相关==========================================================================end

// 通用方法============================================================================begin
func checkIsLegalType(chkType string) bool {
	for _, v := range LegalTypeSlice {
		if chkType == v {
			return true
		}
	}

	return false
}

func getDbFileVersionPostfix(fileName string) string {
	splitStr := fileName
	splitStrVec := strings.Split(fileName, "/")
	splitStrVecLen := len(splitStrVec)
	if splitStrVecLen > 0 {
		splitStr = splitStrVec[splitStrVecLen-1]
	}

	splitVec := strings.Split(splitStr, ".")
	if len(splitVec) > 0 {
		subStr := splitVec[0]
		splitVec2 := strings.Split(subStr, "db_v")
		if len(splitVec2) > 1 {
			return splitVec2[1]
		}
	}

	return ""
}

// 通用方法============================================================================end
