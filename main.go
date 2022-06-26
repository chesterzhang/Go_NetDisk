package main

import (
	"Go_NetDisk/handler"
	"fmt"
	"net/http"
)

func main() {
	//r := mux.NewRouter().Path("/file/upload")
	////r.HandleFunc("/file/upload/{username}",handler.UploadHandler)]
	//r.Host("")

	http.HandleFunc("/file/upload",handler.UploadHandler) // upload 后面要加多一条杠 否则无法识别URL 的参数
	http.HandleFunc("/file/upload/suc",handler.UploadSucHandler)
	http.HandleFunc("/file/meta",handler.GetFileMetaHander)
	http.HandleFunc("/file/query",handler.FileQueryHandler)
	http.HandleFunc("/file/download",handler.DownloadHandler)
	http.HandleFunc("/file/update",handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete",handler.FileDeleteHandler)

	http.HandleFunc("/user/signup",handler.SignupHandler)
	http.HandleFunc("/user/signin",handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))// 多加一层拦截器

	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))// 秒传

	//分块上传接口
	http.HandleFunc("/file/mpupload/init",handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart",handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete",handler.HTTPInterceptor(handler.CompleteUploadHandler))


	// 访问静态资源
	http.Handle("/", http.FileServer(http.Dir("./static")))//http://localhost:8080/view/signin.html
	err:=http.ListenAndServe(":8080",nil)
	if err!=nil {
		fmt.Printf("faile to start server, err: %d",err.Error())
	}

}
