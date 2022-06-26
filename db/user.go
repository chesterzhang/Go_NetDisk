package db

import (
	mydb "Go_NetDisk/db/mysql"
	"fmt"
)

// User : 用户表model
type User struct {
	Username     string
	Email        string
	Phone        string
	SignupDt     string
	LastActiveDt string
	Status       int
}

//注册用户, 成功返回 true, 失败返回 false
func UserSignUp(username string, passwd string) bool  {
	stmt, err := mydb.DBConn().Prepare("INSERT IGNORE INTO tbl_user(`user_name`, `user_pwd`) values(?,?)")
	if err!=nil {
		fmt.Println("Failed to inser, err :" + err.Error())
		return false
	}

	defer stmt.Close()

	ret,err := stmt.Exec(username, passwd)
	if err!=nil {
		fmt.Println("Failed to inser, err :" + err.Error())
		return false
	}

	rowsAffected, err:=ret.RowsAffected()
	if  err==nil && rowsAffected >0 {
		return true
	}
	return  false
	
}

// 判断 账号,密码是否一致
func  UserSignin(username string, encpwd string) bool {
	stmt,err:=mydb.DBConn().Prepare("SELECT * FROM tbl_user where user_name=? LIMIT 1")
	if err!=nil{
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err!=nil{
		fmt.Println(err.Error())
		return false
	}

	if rows== nil{
		fmt.Println("username not found:"+ username)
		return false
	}

	pRows:= mydb.ParseRows(rows)
	if len(pRows)>0 && string(pRows[0]["user_pwd"].([]byte))==encpwd {
		return true
	}

	return false
}

// 刷新用户登录的 token
func UpdateToken(username string, token string )  bool {
	stmt,err:=mydb.DBConn().Prepare("REPLACE INTO tbl_user_token (`user_name`, `user_token`) VALUES (?,?)")
	if err!=nil{
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err =stmt.Exec(username, token)
	if err!=nil{
		fmt.Println(err.Error())
		return false
	}
	return true
}

// GetUserInfo : 查询用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"select user_name,signup_dt from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupDt)
	if err != nil {
		return user, err
	}
	return user, nil
}