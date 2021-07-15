package main

import (
	"log"
	"net"
	"sync"
)

var wg sync.WaitGroup

func read(conn net.Conn) {
	wg.Add(1)
	go func() {
		defer func() {
			log.Println("退出read:", conn.RemoteAddr().String())
			wg.Done()
		}()
		buf := make([]byte, 1024)
		for {
			if conn != nil {
				l, err := conn.Read(buf)
				if err != nil && err.Error() != "EOF" {
					log.Println("读取错误", err)
					return
				}
				if l != 0 {
					log.Println("读取到的数据长度和数据", l, buf[:l])
				}
			}
		}
	}()
}

func main() {
	lister, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Println("lister-error", err)
		return
	}
	conn, err := lister.Accept()
	if err != nil {
		log.Println("err-conn", err)
		return
	}
	log.Println("本次链接地址:", conn.RemoteAddr().String())
	read(conn)
	wg.Wait()
}
