package handler

import (
	rPool "Go_NetDisk/cache/redis"
	dblayer "Go_NetDisk/db"
	"Go_NetDisk/util"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int // 一块的大小
	ChunkCount int // 一共拆分成多少块
}

// InitialMultipartUploadHandler : 初始化分块上传, 将如何分块上传写入 redis
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 2. 获得redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB , 每一小块 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))), // 有多少小块
	}

	// 4. 将初始化信息写入到redis缓存, 数据类型为 hashmap,  HSET 为哈希表中的字段赋值
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)// 多少小块
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash) // 文件hash值
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize) // 文件大小



	// 5. 将响应初始化数据返回到客户端
	//fmt.Println("upinfo", upInfo) //oks
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// UploadPartHandler : 拿到分块上传文件的uploadidx, idx, 然后上传第idx分块的文件
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")// 第几块上传的

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 获得文件句柄，用于存储分块内容
	fpath := "./data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744) // 创建目录, 权限等级, user:7 rwx 可读可写可执行, group:5 r-x, other: 4 r--
	fd, err := os.Create(fpath) // 创建 file description
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024) //1MB
	for {
		n, err := r.Body.Read(buf) // 将 request 的 body 写到 buf 中, 返回 byte 的长度
		fd.Write(buf[:n]) // 将buf 内容写入文件, 也就是将 request 的body 写入文件中
		if err != nil {
			break
		}
	}

	// 4. 更新redis缓存状态, uploadId 的第chkidx个上传的块上传完毕
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler : 查看分块上传是否完成
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid)) // 根据 uploadId 查出对应的 结构体所有字段
	if err != nil { // 如果根据 mu_upid 压根查不到
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}

	totalCount := 0
	chunkCount := 0
	// getall 从 redis 里查出来对应的 uploadId 的 [ 字段0, value0, 字段1, value1,...]
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		// 取到了 uploadId 对应的 chunkcount 字段(也就是总块数), 记录 chunkcount下来
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" { // 如果查到了 chkidx 字段
			chunkCount++ // 说明完成了第个小块的上传
		}
	}
	// 如果全部小块没有上传完成
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// 4. TODO：合并分块

	// 5. 全部分块上传完成, 更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "") //更新 tbl_file 表
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize)) // 更新 tbl_user_file 表

	// 6. 响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}



