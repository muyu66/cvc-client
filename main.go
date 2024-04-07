package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/url"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to TTY Demo")

	var cmdMenu = `
请输入需要执行的编号:
1 对企业评价
2 查看企业综合评分
`

	fmt.Print(cmdMenu)
	input := readInput()

	switch input {
	case "1":
		comment()
	case "2":
		score()
	default:
		fmt.Println("Invalid command. Please enter 1, 2, or 3")
	}
}

func comment() {
	fmt.Print("请输入企业编号:")
	corpIdStr := readInput()
	fmt.Print("请输入分数: (1-10)")
	scoreStr := readInput()
	txCommit(corpIdStr, scoreStr)
	fmt.Println("已提交")
}

func score() {
	fmt.Print("请输入企业编号:")
	corpIdStr := readInput()
	loadComment(corpIdStr)
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func txCommit(corpId string, score string) {
	var key = "Comment/" + corpId + score
	valueJson, err := json.Marshal(map[string]string{
		"corpId": corpId,
		"score":  score,
	})
	value := base64.StdEncoding.EncodeToString(valueJson)
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Set("tx", fmt.Sprintf(`"%s=%s"`, key, value))
	query := params.Encode()

	txCommitUrl := fmt.Sprintf(`http://localhost:26657/broadcast_tx_commit?%s`, query)

	client := resty.New()

	resp, err := client.R().
		Post(txCommitUrl)
	if err != nil {
		fmt.Println("请求发送失败:", err)
		return
	}
	if resp.Error() != nil {
		fmt.Println("请求发送失败:", resp.Error())
		return
	}
	if resp.StatusCode() != 200 {
		fmt.Println("请求发送失败:", string(resp.Body()))
		return
	}
}

func loadComment(corpId string) string {
	var key = "Corp/" + corpId
	params := url.Values{}
	params.Set("data", fmt.Sprintf(`"%s"`, key))
	query := params.Encode()
	loadUrl := fmt.Sprintf(`http://localhost:26657/abci_query?%s`, query)

	client := resty.New()

	resp, err := client.R().
		Get(loadUrl)
	if err != nil {
		log.Fatal("请求发送失败:", err)
	}
	if resp.Error() != nil {
		log.Fatal("请求发送失败:", resp.Error())
	}
	if resp.StatusCode() != 200 {
		log.Fatal("请求发送失败:", string(resp.Body()))
	}

	var result struct {
		Result struct {
			Response struct {
				Value string `json:"value"`
			} `json:"response"`
		} `json:"result"`
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		log.Fatal("解析 JSON 数据时出错:", err)
	}
	value, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		log.Fatal(err)
	}
	return string(value)
}
