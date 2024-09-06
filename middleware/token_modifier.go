package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/buger/goreplay/proto"
	"os"
)

// requestID -> originalToken
// 请求 ID -> 原始 Token
var originalTokens map[string][]byte

// originalToken -> replayedToken
// 原始 Token -> 回放 Token
var tokenAliases map[string][]byte

//var json_data interface{}

func main() {
	originalTokens = make(map[string][]byte)
	tokenAliases = make(map[string][]byte)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		encoded := scanner.Bytes()
		buf := make([]byte, len(encoded)/2)
		hex.Decode(buf, encoded)
		process(buf)
	}
}
func process(buf []byte) {
	// First byte indicate payload type, possible values:
	//  1 - Request
	//  2 - Response
	// 3 - ReplayedResponse
	// 第一个字节表示有效负载类型，可能的值:
	// 1 - 请求
	// 2 - 响应
	// 3 - 回放响应
	payloadType := buf[0]
	headerSize := bytes.IndexByte(buf, '\n') + 1
	header := buf[:headerSize-1]
	// Header contains space separated values of: request type, request id, and request start time (or round-trip time for responses)
	// Header 包含空格分隔的值:请求类型，请求 id，请求开始时间(或响应的往返时间)
	meta := bytes.Split(header, []byte(" "))

	// For each request you should receive 3 payloads (request, response, replayed response) with same request id
	// 对于每个请求，你应该收到 3 个有效负载(request, response, replayed response)，具有相同的请求 id
	reqID := string(meta[1])
	payload := buf[headerSize:]
	Debug("Received payload:", string(buf))
	switch payloadType {
	case '1': // Request
		if bytes.Equal(proto.Path(payload), []byte("/admin/login")) {
			originalTokens[reqID] = []byte{}
			Debug("Found token request:", reqID)
		} else {
			//token, vs, _ := proto.PathParam(payload, []byte("token")) //取到回放响应的 token 值
			token := proto.Header(payload, []byte("token")) //取到原始的 token 值
			Debug("Received token:", string(token))
			if len(token) != 0 { // If there is GET token param
				Debug("If there is GET token param")
				Debug("tokenAliases", tokenAliases)
				if alias, ok := tokenAliases[string(token)]; ok { //检查要替换的 token 值是否存在
					Debug("Received alias")
					// Rewrite original token to alias
					payload = proto.SetHeader(payload, []byte("token"), alias) //将原始的 token 替换成回放的 token
					// Copy modified payload to our buffer
					buf = append(buf[:headerSize], payload...)
				}
			}
		}
		// Emitting data back
		os.Stdout.Write(encode(buf)) //重写请求准备发往回放服务
	case '2': // Original response
		if _, ok := originalTokens[reqID]; ok {
			jsonObject, err := simplejson.NewJson([]byte(proto.Body(payload)))
			if err != nil {
				fmt.Println(err)
			}
			token := jsonObject.Get("token")
			secureToken := token
			f, _ := secureToken.Bytes()
			originalTokens[reqID] = f
			Debug("Remember origial token:", f)
		}
	case '3': // Replayed response
		if originalToken, ok := originalTokens[reqID]; ok {
			delete(originalTokens, reqID)
			//jsonObject, err := simplejson.NewJson([]byte(proto.Body(payload)))
			jsonObject, err := simplejson.NewJson(proto.Body(payload))
			if err != nil {
				fmt.Println("acgsuiuui:", err)
			}
			token := jsonObject.Get("token")
			f, _ := token.Bytes()
			tokenAliases[string(originalToken)] = f //拿到现在的 token 值用来替换掉过去的 token 值
			Debug("Create alias for new token token, was:", string(originalToken), "now:", string(f))
		}
	}
}
func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'
	return dst
}
func Debug(args ...interface{}) {
	if os.Getenv("GOR_TEST") == "" { // if we are not testing
		fmt.Fprint(os.Stderr, "[DEBUG][TOKEN-MOD] ")

		for _, arg := range args {
			if str, ok := arg.(string); ok && str == "Received payload:" {
				fmt.Fprint(os.Stderr, "\x1b[31m") // 设置颜色为红色
				fmt.Fprint(os.Stderr, arg)
				fmt.Fprint(os.Stderr, "\x1b[0m") // 重置颜色
			} else {
				fmt.Fprintln(os.Stderr, args...)
			}
		}
	}
}
