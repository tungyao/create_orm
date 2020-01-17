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
	c        string // manual or auto control
	dataType = map[string]string{"int": "int", "varchar": "string", "timestamp": "string", "bigint": "int64", "tinyint": "int8", "char": "byte", "text": "string", "float": "float32", "double": "float64"}
)

func init() {
	flag.StringVar(&t, "t", "user", "Name of the mapping table,this column is must first")
	flag.StringVar(&s, "s", "", "Name of the mapping file")
	flag.StringVar(&d, "d", "", "Name of the database")
	flag.StringVar(&u, "u", "", "Name of the login name")
	flag.StringVar(&p, "p", "", "Name of the login password")
	flag.StringVar(&f, "f", "./a.go", "file path")
	flag.StringVar(&c, "c", "auto", "manual input or auto input")
}
func main() {
	flag.Parse()
	fs, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE, 777)
	switch c {
	case "manual":
		goto manual
	case "auto":
		goto auto
	}
manual:
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	fs.Write([]byte("\n// **************" + s + " Start************\n"))
	fs.Write([]byte("type " + s + " struct {\n"))
	for _, v := range flag.Args() {
		arg := SplitString([]byte(v), []byte(":"))
		fs.Write([]byte("\t" + string(arg[0]) + "\t" + string(arg[1]) + "\n"))
	}
	fs.Write([]byte("}\n\n"))
	fs.Write([]byte("func Get" + s + "Struct() *" + s + " {\n\treturn new(" + s + ")\n}\n"))
	fs.Write([]byte("// --------------" + s + " End--------------\n"))
	fs.Close()
	os.Exit(0)
auto:
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
	if len(flag.Args()) > 0 {
		index := ""
		for _, v := range flag.Args() {
			index += v + ","
		}
		index = index[:len(index)-1]
		rows, err := D.Query("select " + index + " from " + t + " limit 1")
		if err != nil {
			log.Println(err)
		}
		column, _ := rows.Columns()
		fmt.Println(column)
		os.Exit(0)
	}
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
	fs, err = os.OpenFile(f, os.O_APPEND|os.O_CREATE, 777)
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
func SplitString(str []byte, p []byte) [][]byte {
	group := make([][]byte, 0)
	ps := 0
	for i := 0; i < len(str); i++ {
		if str[i] == p[0] && i < len(str)-len(p) {
			if len(p) == 1 {
				group = append(group, str[ps:i])
				ps = i + len(p)
				//return [][]byte{str[:i], str[i+1:]}
			} else {
				for j := 1; j < len(p); j++ {
					if str[i+j] != p[j] || j != len(p)-1 {
						continue
					} else {
						group = append(group, str[ps:i])
						ps = i + len(p)
					}
					//return [][]byte{str[:i], str[i+len(p):]}
				}
			}
		} else {
			continue
		}
	}
	group = append(group, str[ps:])
	return group
}
