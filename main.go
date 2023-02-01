package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

var (
	cosClient *cos.Client
)

const (
	ApnicURL          = "http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest"
	ApnicCNIPv4Prefix = "apnic|CN|ipv4"
	ApnicIPv6Prefix   = `apnic\|\w{2}\|ipv6`
	ObjectName        = "apnic-cn"
)

func parseLine2IPNet(line string) (*net.IPNet, error) {
	items := strings.Split(line, "|")
	if len(items) < 4 {
		return nil, fmt.Errorf("parse ipnet items less 4: %s", items)
	}
	allocatedCount, err := strconv.Atoi(items[4])
	if err != nil {
		return nil, err
	}
	_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%v", items[3], 32-math.Log2(float64(allocatedCount))))
	if err != nil {
		return nil, err
	}
	return ipNet, nil
}

func parseApnicCN2Cos() {
	// 获取 Apnic 内容
	resp, err := http.Get(ApnicURL)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("http get apnic err: " + err.Error())
		return
	}
	fmt.Printf("http get resp code: %d\n", resp.StatusCode)
	reader := bufio.NewReader(resp.Body)
	// 解析出 CN IP
	buf := new(bytes.Buffer)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("reader read string err: " + err.Error())
			break
		}
		if ok, _ := regexp.MatchString(ApnicIPv6Prefix, line); ok {
			break
		}
		if strings.HasPrefix(line, ApnicCNIPv4Prefix) {
			ipNet, err := parseLine2IPNet(line)
			if err != nil {
				fmt.Printf("parse ipnet %v err: %v\n", line, err)
			}
			buf.WriteString(ipNet.String() + "\n")
		}
	}
	_, err = cosClient.Object.Put(context.Background(), ObjectName, buf, nil)
	if err != nil {
		fmt.Println("upload to cos err: " + err.Error())
	}
}

func init() {
	// 初始化 cos 客户端
	u, _ := url.Parse(os.Getenv("COS_OBJECT_URL"))
	b := &cos.BaseURL{BucketURL: u}
	cosClient = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			// 通过环境变量获取密钥
			// 环境变量 COS_SECRETID 表示用户的 SecretId，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretID: os.Getenv("COS_SECRETID"),
			//// 环境变量 COS_SECRETKEY 表示用户的 SecretKey，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretKey: os.Getenv("COS_SECRETKEY"),
			// Debug 模式，把对应 请求头部、请求内容、响应头部、响应内容 输出到标准输出
			Transport: &debug.DebugRequestTransport{
				RequestHeader: true,
				// Notice when put a large file and set need the request body, might happend out of memory error.
				RequestBody:    false,
				ResponseHeader: true,
				ResponseBody:   false,
			},
		},
	})

}

func main() {
	// Make the handler available for Remote Procedure Call by Cloud Function
	cloudfunction.Start(parseApnicCN2Cos)
}
