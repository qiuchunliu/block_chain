package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(genesisBlock)
		BlockChain = append(BlockChain, genesisBlock)
	}()
	log.Fatal(run())
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

// 通常来说，更长的链表示它的数据（状态）是比较新的
// 所以我们需要一个函数能帮我们将本地的过期的链切换成最新的链
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(BlockChain) {
		BlockChain = newBlocks
	}
}

/*
	我们基本就把所有重要的函数完成了。
	接下来，我们需要一个方便直观的方式来查看我们的链，包括数据及状态。
	通过浏览器查看 web 页面可能是最合适的方式
*/
// 借助 Gorilla/mux 包，我们先写一个函数来初始化我们的 web 服务
func handleGetBlockChain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(BlockChain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

type Message struct {
	BPM int
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(BlockChain[len(BlockChain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, BlockChain[len(BlockChain)-1]) {
		newBlockChain := append(BlockChain, newBlock)
		replaceChain(newBlockChain)
		spew.Dump(BlockChain)
	}
	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockChain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listen on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
