package handler

import (
	dblayer "Go_NetDisk/db"
	"Go_NetDisk/meta"
	"Go_NetDisk/store/oss"
	"Go_NetDisk/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 文件元信息

type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadDt string
}

func UploadHandler(w http.ResponseWriter, r * http.Request)  {

	if r.Method  == "GET"{
		//fmt.Println("enter here")
		data, err:=ioutil.ReadFile("./static/view/index.html")// 加载HTML页面
		if err!=nil {
			io.WriteString(w,"internal server error")
			return
		}
		io.WriteString(w,string(data))
	}else if r.Method=="POST" {

		//接收前端上传来的文件, FormFile 参数与前端的 id 相同
		file, head, err:= r.FormFile("file")
		if err!=nil {
			 fmt.Printf("failed to get data, err:%s \n", err.Error())
			return
		}

		defer file.Close()

		//创建 文件元信息
		fileMeta:= meta.FileMeta{
			FileName: head.Filename,
			Location: "./temp/" + head.Filename,
			UploadDt: time.Now().Format("2006-01-02 15:04:05"),
		}

		// 将文件写入本地存储
		newFile, err :=os.Create(fileMeta.Location)
		if err!=nil {
			fmt.Printf("failed to create file, err:%s \n", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err  = io.Copy(newFile,file)
		if err!=nil {
			fmt.Printf("failed to save data into file, err:%s \n", err.Error())
			return
		}

		// 继续完善文件元信息
		newFile.Seek(0,0) // 方法用于移动文件读取指针到指定位置
		fileMeta.FileSha1 = util.FileSha1(newFile) // 获取 hash 值
		newFile.Seek(0,0) // 方法用于移动文件读取指针到指定位置

		// 将文件写入 oss
		ossPath := "oss/" + fileMeta.FileSha1
		err = oss.Bucket().PutObject(ossPath, newFile)
		if err != nil {
			fmt.Println(err.Error())
			w.Write([]byte("Upload failed!"))
			return
		}
		fileMeta.Location = ossPath


		// 持久化数据, 将文件元信息 存储到mysql中
		suc :=meta.UpdateFileMetaDB(fileMeta)
		if !suc {
			w.Write([]byte("Faile to update file in tbl_file."))
		}

		// 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")

		suc = dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/view/home.html", http.StatusFound)// 上传成功后又返回主页
		} else {
			w.Write([]byte("Faile to update file in tbl_user_file."))
		}

	}
}

func UploadSucHandler(w http.ResponseWriter, r * http.Request){
	io.WriteString(w,"Upload finished! ")
}

// 获取文件元信息
func GetFileMetaHander(w http.ResponseWriter, r * http.Request)  {
	r.ParseForm()

	filehash:=r.Form["filehash"][0]

	//fMeta:=meta.GetFileMeta(filehash)// 从 fileMetas[string] FileMeta 里面获取 文件元信息
	fMeta, err := meta.GetFileMetaDB(filehash)// 从 Mysql 里面获取 文件元信息
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data,err:=json.Marshal(fMeta)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 获取 指定个数的文件, 按照上传时间先后顺序
func FileQueryHandler(w http.ResponseWriter, r * http.Request){
	//fmt.Println("enter FileQueryHandler") //ok here
	r.ParseForm()
	limitCnt,_:= strconv.Atoi(r.Form.Get("limit")) // 解析 limit 参数并转化为 int 类型
	//fileMetas:=meta.GetLastFileMetas(limitCnt)

	username := r.Form.Get("username")
	userFiles,err :=dblayer.QueryUserFileMetas(username,limitCnt)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Println(userFiles) // 这里是没有问题的

	data, err:=json.Marshal(userFiles)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}


// 根据URL里面的 filehash 参数,也就是文件的 hashcode 获得文件, 然后下载
func DownloadHandler(w http.ResponseWriter, r * http.Request)  {
	r.ParseForm()
	fsha1:=r.Form.Get("filehash")// 获取参数
	fm:=meta.GetFileMeta(fsha1)
	f,err:=os.Open(fm.Location)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer  f.Close()
	data,err:=ioutil.ReadAll(f)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/octect-stream")
	w.Header().Set("content-disposition","attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)
}


// 更新元信息接口, 重命名
func FileMetaUpdateHandler(w http.ResponseWriter, r * http.Request)  {
	r.ParseForm()
	opType:=r.Form.Get("op")
	fileSha1:=r.Form.Get("filehash")
	newFileName:=r.Form.Get("filename")

	if opType!="0" {// 目前只支持重命名操作
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method!="POST" {// 目前只支持重命名操作
		w.WriteHeader(http.StatusForbidden)
		return
	}

	curFileMeta:=meta.GetFileMeta(fileSha1)
	curFileMeta.FileName=newFileName
	meta.UpdateFileMeta(curFileMeta)

	data,err:=json.Marshal(curFileMeta)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// 删除文件与文件元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()

	fileSha1:=r.Form.Get("filehash")
	fMeta:=meta.GetFileMeta(fileSha1)

	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(fileSha1)


	w.WriteHeader(http.StatusOK)
}


// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	//fmt.Println("1111") //到这里
	// 2. 从文件表中查询相同hash的文件记录
	//fileMeta, err := meta.GetFileMetaDB(filehash)//
	_, err := meta.GetFileMetaDB(filehash) // 尝试查找文件的 hash , 如果hash 都没有, 那么肯定之前没有传过
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 3. 查不到记录则返回秒传失败
	//if  filehash != fileMeta.FileSha1     {
	//if   fileMeta.FileSha1 ==""    {
	//	resp := util.RespMsg{
	//		Code: -1,
	//		Msg:  "秒传失败，请访问普通上传接口",
	//	}
	//	w.Write(resp.JSONBytes())
	//	return
	//}

	//fmt.Println("here") //到不了

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

// DownloadURLHandler:  生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	row, _ := dblayer.GetFileMeta(filehash)
	signedURL := oss.DownloadURL(row.FileAddr.String)
	w.Write([]byte(signedURL))
}

