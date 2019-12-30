package test

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"testing"
)

type User struct {
	Id         int
	Name       string
	Password   string
	CreateTime string
}

func TestReady(t *testing.T) {
	DB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%s)/%s", "root", "1121331", "tcp", "localhost", "3306", "test"))
	if err != nil {
		log.Println(err)
	}
	//row, _ := DB.Query("select * from user where id=?", 1)
	row, err := DB.Query(`select COLUMN_NAME,DATA_TYPE,IS_NULLABLE,TABLE_NAME,COLUMN_COMMENT from information_schema.COLUMNS where table_schema=? and table_name=?`, "test", "user")
	if err != nil {
		log.Println(err)
	}
	col, _ := row.Columns()
	tc := make([]map[string]string, 0)
	for row.Next() {
		value := make([]interface{}, len(col))
		columnPointers := make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			columnPointers[i] = &value[i]
		}
		err = row.Scan(columnPointers...)
		if err != nil {
			log.Println(err)
		}
		data := make(map[string]string)
		for i := 0; i < len(col); i++ {
			columnValue := columnPointers[i].(*interface{})
			data[col[i]] = string((*columnValue).([]uint8))
		}
		tc = append(tc, data)
	}
}
