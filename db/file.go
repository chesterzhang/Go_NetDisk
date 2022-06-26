package db

import (
	mydb "Go_NetDisk/db/mysql"
	"database/sql"
	"fmt"
)

// OnFileUploadFinished : 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	//conn:=mydb.DBConn()
	//if conn==nil {
	//	fmt.Print("nil pointer here")
	//}
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before \n", filehash)
		}
		return true
	}
	return false
}

type  TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// 根据 filesha1 从数据库里面查询 文件元信息
func GetFileMeta( filehash string) (*TableFile , error)  {
	stmt, err := mydb.DBConn().Prepare(
		" SELECT file_sha1, file_addr, file_name, file_size FROM tbl_file" +
			" WHERE file_sha1=? AND status=1 LIMIT 1 ")

	if err !=nil {
		fmt.Print(err.Error())
		return nil,err
	}
	
	defer stmt.Close()
	tfile:=TableFile{}
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err !=nil {
		fmt.Print(err.Error())
		return nil,err
	}
	return &tfile, nil
}
