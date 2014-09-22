// kindBulk
//   没有：nil（没有这个 key）
//   有  ：string（有这个 key）

// kindMultiBulk
//   没有：[]interface{}{}（没有这个 key，len() == 0）
//   有  ：[]interface{}{"str1", "str2", "str3"}（全部子 key 都有，len() == len(subKeys)）
//   有洞：[]interface{}{"str1", nil, "str3"}（有些子 key 没有，len() < len(subKeys)）
//   超时：nil（目前只有 BLPOP 会返回这个值？）

/*redis返回数据
Error: 			-Error message\r\n

Simple String: 	+OK\r\n

Integers: 		:1000\r\n(int64)

Bulk Strings: 	$6\r\nfoobar\r\n
				$0\r\n\r\n
				$-1\r\n

Arrays:			*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
				*0\r\n
				*-1\r\n
A client sends to the Redis server a RESP Array consisting of just Bulk Strings.
A Redis server replies to clients sending any valid RESP data type as reply.
*/
package msgredis

// func (c *Conn) writeInt1(n int) error {
// 	p := []byte{'$'}
// 	// 调用两次itoa，且还要对string转回byte
// 	nstr := strconv.Itoa(n)
// 	p = append(p, strconv.Itoa(len(nstr))...)
// 	p = append(p, nstr...)
// 	p = append(p, '\r', '\n')
// 	_, e := c.wb.Write(p)
// 	if e != nil {
// 		return e
// 	}
// }

// painc recover

// 1 网络错误与普通错误的区分
// 2 redis命令参数，统一为一个结构体？
// 3 nil与空的处理
// 4 可能出现panic的处理
// 5 读到了错误字符，会不会有没有读完的数据

// song
// when we stand together
