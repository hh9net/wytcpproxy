package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

var lock sync.Mutex
var trueList []string
var trustList []string
var ip string
var list string
var bmd string

func main() {
	flag.StringVar(&ip, "l", ":9897", "-l=0.0.0.0:9897 指定服务监听的端口")
	flag.StringVar(&list, "d", "127.0.0.1:1789,127.0.0.1:1788", "指定后端的IP和端口,多个用','隔开")
	flag.StringVar(&bmd, "b", "127.0.0.2,127.0.0.1", "指定白名单的IP,多个用','隔开")
	flag.Parse()
	trueList = strings.Split(list, ",")
	trustList = strings.Split(bmd, ",")
	if len(trueList) <= 0 {
		fmt.Println("后端IP和端口不能空,或者无效")
		os.Exit(1)
	}
	if len(trustList) <= 0 {
		fmt.Println("白名单不能为空,不然都访问不了了啊")
		os.Exit(1)
	}
	server()
}

func In(k []string, e string) bool {
	for _, d := range k {
		if e == d {
			return true
		}
	}
	return false

}

func server() {
	lis, err := net.Listen("tcp", ip)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("建立连接错误:%v\n", err)
			continue
		}
		fmt.Println(conn.RemoteAddr(), conn.LocalAddr())
		if len(conn.RemoteAddr().String()) > 0 {
			ipport := strings.Split(conn.RemoteAddr().String(), ":")
			if !In(trustList, ipport[0]) {
				fmt.Println("你的ip不允许")
				continue
			}
		}
		go handle(conn)
	}
}

func handle(sconn net.Conn) {
	defer sconn.Close()
	ip, ok := getIP()
	if !ok {
		return
	}
	dconn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", ip, err)
		return
	}
	ExitChan := make(chan bool, 1)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(dconn, sconn)
		fmt.Printf("往%v发送数据失败:%v\n", ip, err)
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	go func(sconn net.Conn, dconn net.Conn, Exit chan bool) {
		_, err := io.Copy(sconn, dconn)
		fmt.Printf("从%v接收数据失败:%v\n", ip, err)
		ExitChan <- true
	}(sconn, dconn, ExitChan)
	<-ExitChan
	dconn.Close()
}

func getIP() (string, bool) {
	lock.Lock()
	defer lock.Unlock()

	if len(trueList) < 1 {
		return "", false
	}
	ip := trueList[0]
	trueList = append(trueList[1:], ip)
	return ip, true
}
