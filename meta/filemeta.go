package meta

import (
	mydb "Go_NetDisk/db"
	"sort"
)


// 文件元信息
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadDt string
}

// 通过 FileSha1 查找 文件的 文件元信息
var fileMetas map[string] FileMeta

func init()  {
	fileMetas=make(map[string] FileMeta)
}

// 新增/更新 文件元信息
func UpdateFileMeta( fmeta FileMeta)  {
	fileMetas[fmeta.FileSha1]=fmeta
}

// 新增/ 更新 文件元信息到 mysql 中
func  UpdateFileMetaDB(fmeta FileMeta) bool{
	return mydb.OnFileUploadFinished(fmeta.FileSha1,fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

//获取文件的元信息对象
func GetFileMeta( fileSha1 string) FileMeta {
	return  fileMetas[fileSha1]
}

// 从数据库里面获取文件元信息
func GetFileMetaDB( fileSha1 string) (FileMeta,error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if err!=nil{
		return FileMeta{},err
	}
	fmeta:=FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
		UploadDt: "",
	}
	return  fmeta,nil
}

// GetLastFileMetas : 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}

	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// 删除 文件元信息
func RemoveFileMeta(fileSha1 string)  {
	delete(fileMetas, fileSha1)
}
