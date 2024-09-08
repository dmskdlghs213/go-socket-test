package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// シグナル以外でプロセス全体を終了させたい場合はWithCancelを使う
	ctx := context.Background()
	// シグナル起点でDone()が呼ばれるctxを生成
	ctx, cancel := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
		syscall.SIGKILL,
	)
	defer cancel()

	srv, err := newServer()
	if err != nil {
		panic(err)
	}
	// サーバの起動
	go srv.serve(ctx)

	// シグナルごとに処理が変わる場合はswitch
	// シグナル起点でDone()が呼ばれたら処理中のプロセスを待機して終了とする
	<-ctx.Done()
	fmt.Println("server shutdown start")
	// シャットダウン処理
	srv.shutdown()
	fmt.Println("server shutdown complete")
}

func (s *server) shutdown() {
	s.listener.Close()
	s.wg.Wait()
}

type server struct {
	listener *net.TCPListener
	wg       sync.WaitGroup
}

func newServer() (*server, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return &server{
		listener: l,
	}, nil
}

func (s *server) serve(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from:%v\n", r)
		}
		s.listener.Close()
	}()
	fmt.Println("start to tcp server :8080")
	var tempDelay time.Duration
LOOP:
	for {
		select {
		case <-ctx.Done(): // シグナルのctx.Done()が来たらreturnして終了
			return
		default:
			conn, err := s.listener.AcceptTCP()
			if err != nil {
				// タイムアウトエラーの場合は時間を遅延しながらリトライ
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					tempDelay *= 5
					time.Sleep(tempDelay)
					if tempDelay > 1*time.Second { // 1秒以上の場合はreturn error
						return
					}
					continue LOOP
				}
				return
			}
			conn.SetDeadline(time.Now().Add(30 * time.Second)) // TCPコネクションのタイムアウト設定
			s.wg.Add(1)                                        // wg.Add(1)でgoroutineの数をカウント
			go func() {
				s.handleConnection(ctx, conn)
				s.wg.Done() // wg.Done()でgoroutineの数をデクリメント
			}()
		}
	}
}

func (s *server) handleConnection(ctx context.Context, conn *net.TCPConn) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from:%v\n", r)
		}
		conn.Close()
	}()
	// TCPコネクションのタイムアウト設定
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	fmt.Printf("conn read start: %s\n", conn.RemoteAddr())
LOOP:
	for {
		select {
		// シグナルのctx.Done()が来たらreturnして終了
		case <-ctx.Done():
			return
		default:
			request := make([]byte, 100)
			_, err := conn.Read(request)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok &&
					opErr.Timeout() {
					continue LOOP
				} else if err != io.EOF {
					// EOFはクライアントからの読み取りが終了した合図なのでreturnして終了
					return
				}
				// need err log
				return
			}
			// fmt.Printf("received message: %s\n", string(request))
			fmt.Printf("conn read end: %s\n", conn.RemoteAddr())
			response := []byte("HELLO_RESPONSE")
			_, err = conn.Write(response)
			if err != nil {
				// need err log
				return
			}
		}
	}
}
