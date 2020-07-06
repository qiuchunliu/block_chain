package main

import "fmt"

type Block struct {
	Index     int    // 这个块在整个链中的位置
	TimeStamp string // 块生成时的时间戳
	BPM       int    // 每分钟的心跳次数，即心率
	Hash      string // 这个块通过SHA256生成的散列值
	PreHash   string // 前一个块的SHA256散列值
}

func (block *Block) getRecord() string {
	return block.PreHash

}
func main() {
	var b Block
	b.PreHash = "aaa"
	fmt.Println(b.getRecord())
}
