package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

var (
	t        string // table name
	s        string // struct name
	d        string // database name
	u        string // mysql login name
	p        string // mysql login password
	f        string // file path
	dataType = map[string]string{"int": "int", "varchar": "string", "timestamp": "string", "bigint": "int64", "tinyint": "int8", "char": "byte", "text": "string", "float": "float32", "double": "float64"}
)

func init() {
	flag.StringVar(&t, "t", "user", "Name of the mapping table,this column is must first")
	flag.StringVar(&s, "s", "", "Name of the mapping file")
	flag.StringVar(&d, "d", "", "Name of the database")
	flag.StringVar(&u, "u", "", "Name of the login name")
	flag.StringVar(&p, "p", "", "Name of the login password")
	flag.StringVar(&f, "f", "./a.go", "file path")
}
func main() {
	flag.Parse()
	if t == "" {
		fmt.Println("that must have table name")
		os.Exit(0)
	}
	if u == "" {
		fmt.Println("login name have none")
		os.Exit(0)
	}
	if p == "" {
		fmt.Println("login password have none")
		os.Exit(0)
	}
	D, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%s)/%s", u, p, "tcp", "localhost", "3306", d))
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	row, err := D.Query(`select COLUMN_NAME,DATA_TYPE,IS_NULLABLE,TABLE_NAME,COLUMN_COMMENT from information_schema.COLUMNS where table_schema=? and table_name=?`, d, t)
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
	// now ,we got table struct => tc
	fs, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE, 777)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	fs.Write([]byte("\n// **************" + s + " Start************\n"))
	fs.Write([]byte("type " + s + " struct {\n"))
	for _, v := range tc {
		fs.Write([]byte("\t" + string(strings.ToUpper(string(v["COLUMN_NAME"][0]))) + string(v["COLUMN_NAME"][1:]) + "\t" + dataType[v["DATA_TYPE"]] + "\t`" + fmt.Sprint(v["COLUMN_COMMENT"]) + "`\n"))
	}
	fs.Write([]byte("}\n\n"))
	fs.Write([]byte("func Get" + s + "Struct() *" + s + " {\n\treturn new(" + s + ")\n}\n"))
	fs.Write([]byte("// --------------" + s + " End--------------\n"))
	fs.Close()
}
