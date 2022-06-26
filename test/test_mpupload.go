package main

import (
	"bufio"
	"bytes"
	"fmt"
	jsonit "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// 拿到文件路径, 按照一个个chunksize写入一个buffer, 将
func multipartUpload(filename string, targetURL string, chunkSize int) error {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f) // 返回一个reader
	index := 0

	ch := make(chan int) // 用于传输 第几个部分已经上传好了
	buf := make([]byte, chunkSize) //每次读取chunkSize大小的内容, 这里是 5MB
	for {
		n, err := bfRd.Read(buf) // 将文件的chunkSize读到 buf 中去, 返回读到的长度

		//如果已经读完了, 跳出循环
		if n <= 0 {
			break
		}
		index++

		bufCopied := make([]byte, 5*1048576) //5MB
		copy(bufCopied, buf)

		go func(b []byte, curIdx int) {
			fmt.Printf("upload_size: %d\n", len(b))
			// 请求 UploadPartHandler : 拿到分块上传文件的uploadidx, idx, 然后上传第idx分块的文件,将 第idx块上传成功写入 redis
			resp, err := http.Post(
				targetURL+"&index="+strconv.Itoa(curIdx),
				"multipart/form-data", //用以支持文件上传
				bytes.NewReader(b))// 分块文件
			if err != nil {
				fmt.Println(err)
			}

			body, er := ioutil.ReadAll(resp.Body) 	// body 应该为null
			fmt.Printf("%+v %+v\n", string(body), er)
			resp.Body.Close()

			ch <- curIdx
		}(buf[:n], index) //(bufCopied[:n], index)

		//遇到任何错误立即返回，并忽略 EOF 错误信息
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
			}
		}
	}

	for idx := 0; idx < index; idx++ {
		select {
		case res := <-ch:
			fmt.Println(res)
		}
	}

	return nil
}

func main() {
	username := "admin"
	token := "79d63abb423d04e3cbeb204143a1047662b47756" //admin&token=79d63abb423d04e3cbeb204143a1047662b47756
	filehash := "ec3d25dfc6e199621b79d992c50b7d87f0774c3c"

	// 1. 请求初始化分块上传接口, 将 filehash, filesize, chunkcount 写入 redis
	// response body 中含有 MultipartUploadInfo 结构体
	//type MultipartUploadInfo struct {
	//	FileHash   string
	//	FileSize   int
	//	UploadID   string
	//	ChunkSize  int // 一块的大小
	//	ChunkCount int // 一共拆分成多少块
	//}
	resp, err := http.PostForm(
		"http://localhost:8080/file/mpupload/init",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {"16455073"},
		})
	fmt.Println(resp.StatusCode)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) //
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Println("body",body)

	//2. 得到uploadID以及服务端指定的分块大小chunkSize
	uploadID := jsonit.Get(body, "data").Get("UploadID").ToString()
	chunkSize := jsonit.Get(body, "data").Get("ChunkSize").ToInt()
	fmt.Printf("uploadid: %s  chunksize: %d\n", uploadID, chunkSize)

	// 3. 请求分块上传接口
	filename := "CV_template.rar"
	tURL := "http://localhost:8080/file/mpupload/uppart?" +
		"username=admin&token=" + token + "&uploadid=" + uploadID
	multipartUpload(filename, tURL, chunkSize)

	// 4. 请求分块完成接口
	resp, err = http.PostForm(
		"http://localhost:8080/file/mpupload/complete",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {"132489256"},
			"filename": {"CV_template.rar"},
			"uploadid": {uploadID},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Printf("complete result: %s\n", string(body))
}
