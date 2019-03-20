package queue


import (
	"fmt"
	"time"
	"math/rand"
	"encoding/binary"
)

/*
网络协议规定为：
index(byte) size(two bytes) content ...
*/
var (
	Req_REGISTER byte = 1 // 1 --- c register cid
	Res_CURRENT_POS byte = 2 // 2 --- s response current queue position
	Res_GET_TOKEN byte = 3   // 3 --- s response token
)



func init() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("init called")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}


func AssemblePacket(index uint8, content []byte) []byte {
	head := []byte{index}
	size := make([]byte, 2)
	binary.LittleEndian.PutUint16(size, uint16(len(content)))
	return append(append(head[:], size[:]...), content[:]...)
}

func DisassemblePacket(data []byte) ([]byte, uint8, []byte) {
	head := uint8(data[0])
	size := binary.LittleEndian.Uint16(data[1:3])

	if len(data) >= int(3 + size) {
		content := data[3:3+size]
		return data[3+size:], head, content
	}
	return data, 0, nil
}

func Int32ToBytes(i int32) []byte {
    var buf = make([]byte, 4)
    binary.LittleEndian.PutUint32(buf, uint32(i))
    return buf
}

func BytesToInt32(buf []byte) int32 {
    return int32(binary.LittleEndian.Uint32(buf))
}

/*
func main() {
	data := append(AssemblePacket(1, []byte("asdfasdfaasdfsa")), AssemblePacket(2, []byte("fdddddddddf"))...)
	fmt.Println("data: ", data)

	data, index, content := DisassemblePacket(data)
	fmt.Println(index, "： ", string(content))

	data, index, content = DisassemblePacket(data)
	fmt.Println(index, "： ", string(content))
}	
*/