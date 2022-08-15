package mq

import  cmn "Go_NetDisk/common"

// 消息队列的 data 格式
type TransferData struct {
	FileHash      string // 文件 hash
	CurLocation   string // 文件当前位置
	DestLocation  string // 文件目标位置
	DestStoreType cmn.StoreType //
}
