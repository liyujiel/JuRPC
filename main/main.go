package main

import (
	"encoding/json"
	"fmt"
	"jurpc"
	"jurpc/codec"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error", err)
	}

	log.Println("start rpc server on", l.Addr())

	addr <- l.Addr().String()
	jurpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	connect, _ := net.Dial("tcp", <-addr)
	defer func() {
		connect.Close()
	}()

	time.Sleep(time.Second)

	_ = json.NewEncoder(connect).Encode(jurpc.DefaultOption)
	cc := codec.NewGobCodec(connect)

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		s := fmt.Sprintf("jurpc request %d", h.Seq)
		_ = cc.Write(h, &s)
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
