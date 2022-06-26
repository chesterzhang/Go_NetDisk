package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var db* sql.DB // 连接对象

func Init()  {
	//db, err :=sql.Open("mysql", "root:123456@tcp(172.17.0.2:3301)/fileserver?charset=utf8")
	db, err :=sql.Open("mysql", "root:123456@tcp(localhost:3301)/fileserver?charset=utf8")
	if err!=nil{
		fmt.Println("Failed to connect to mysql, err: "+ err.Error())
		os.Exit(1)
	}

	//db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)

	err =db.Ping()
	if err!=nil{
		fmt.Println("Failed to connect to mysql, err: "+ err.Error())
		os.Exit(1)
	}

}

func DBConn() *sql.DB{
	//db, err :=sql.Open("mysql", "root:123456@tcp(172.17.0.2:3301)/fileserver?charset=utf8")
	db, err :=sql.Open("mysql", "root:123456@tcp(localhost:3301)/fileserver?charset=utf8")
	if err!=nil{
		fmt.Println("Failed to connect to mysql, err: "+ err.Error())
		os.Exit(1)
	}

	//db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)

	err =db.Ping()
	if err!=nil{
		fmt.Println("Failed to connect to mysql, err: "+ err.Error())
		os.Exit(1)
	}
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}