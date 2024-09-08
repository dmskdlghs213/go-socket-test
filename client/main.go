package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		fmt.Println("request_count: ", i)
		go sendSocketWithWG(wg)
	}
	wg.Wait()
}

func sendSocketWithWG(wg *sync.WaitGroup) {
	defer wg.Done()
	message := "HELLO SOCKET SERVER !!!"
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(7 * time.Second)
	conn.Write([]byte(message))
	response := make([]byte, 1000_000)
	readLen, err := conn.Read(response)
	if err != nil {
		log.Println("read error")
	}
	if readLen == 0 {
		log.Println("zero read")
	}
	fmt.Println(string(response))
}

func sendSocket() {
	message := "HELLO_REQUEST"

	serverIP := "0.0.0.0"
	serverPort := "8080"
	conn, err := net.DialTimeout("tcp", serverIP+":"+serverPort, 10*time.Millisecond)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	time.Sleep(3 * time.Second)

	// メッセージをバイトに変換して送信
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn.Write([]byte(message))

	// サーバからのレスポンスがくるまで待機
	response := make([]byte, 1000_000)
	readLen, _ := conn.Read(response)

	if readLen == 0 {
		log.Println("zero read")
	}
	fmt.Println(string(response))

}
