package main

import (
	//"encoding/hex"
	//"encoding/json"
	//"io"
	//"log"
	//"net/http"
	//"os"
	//"time"

	//"github.com/davecgh/go-spew/spew"
	//"github.com/gorilla/mux"
	//"github.com/joho/godotenv"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now())
}

// 组成区块链的每个块的数据模型
type Block struct {
	Index     int    // 这个块在整个链中的位置
	TimeStamp string // 块生成时的时间戳
	BPM       int    // 每分钟的心跳次数，即心率
	Hash      string // 这个块通过SHA256生成的散列值
	PreHash   string // 前一个块的SHA256散列值
}

func (block *Block) getRecord() string {
	return string(block.Index) + block.TimeStamp + string(block.BPM) + block.PreHash
}

// 定义一个结构表示整个链，最简单的表示形式就是一个 Block 的 slice
var BlockChain []Block

/*
	我们使用散列算法（SHA256）来确定和维护链中块和块正确的顺序
	确保每一个块的 PrevHash 值等于前一个块中的 Hash 值
	这样就以正确的块顺序构建出链：
*/
/*
	我们为什么需要散列？主要是两个原因：
		在节省空间的前提下去唯一标识数据。
			散列是用整个块的数据计算得出，在我们的例子中，将整个块的数据通过 SHA256 计算成一个定长不可伪造的字符串。
		维持链的完整性。
			通过存储前一个块的散列值，我们就能够确保每个块在链中的正确顺序。
			任何对数据的篡改都将改变散列值，同时也就破坏了链。
	以我们从事的医疗健康领域为例，比如有一个恶意的第三方为了调整“人寿险”的价格，
	而修改了一个或若干个块中的代表不健康的 BPM 值，那么整个链都变得不可信了。
*/

// 计算给定的数据的 SHA256 散列值：
func calculateHash(block Block) string {
	//record := string(block.Index) + block.TimeStamp + string(block.BPM) + block.PreHash
	record := block.getRecord()
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// 生成块
/*
	其中，Index 是从给定的前一块的 Index 递增得出
	时间戳是直接通过 time.Now() 函数来获得的
	Hash 值通过前面的 calculateHash 函数计算得出
	PrevHash 则是给定的前一个块的 Hash 值
*/
func generateBlock(oldBlock Block, BPM int) (Block, error) {
	var newBlock Block

	newBlock.Index = oldBlock.Index + 1
	newBlock.TimeStamp = time.Now().String()
	newBlock.BPM = BPM
	newBlock.PreHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)
	return newBlock, nil
}

// 校验块
/*
	校验一个块是否有被篡改：
	检查 Index 来看这个块是否正确得递增
	检查 PrevHash 与前一个块的 Hash 是否一致
	通过 calculateHash 检查当前块的 Hash 值是否正确
*/
func isBlockValid(newBlock, oldBlock Block) bool {
	if (oldBlock.Index + 1) != newBlock.Index {
		return false
	}
	if oldBlock.Hash != newBlock.PreHash {
		return false
	}
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

/*
	除了校验块以外，我们还会遇到一个问题：
	两个节点都生成块并添加到各自的链上，那我们应该以谁为准？
	记住一个原则：始终选择最长的链：
*/
