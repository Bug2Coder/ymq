package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// 0-15的报文类型值
const (
	// RESERVED 是保留值，应视为无效消息类型
	RESERVED byte = iota

	// CONNECT 客户端请求连接服务器版本：v3.1、v3.1.1、v5.0 目录：客户端到服务器
	CONNECT

	// CONNACK 接受确认版本：v3.1、v3.1.1、v5.0 目录：服务器到客户端
	CONNACK

	// PUBLISH 发布客户端到服务器，或服务器到客户端。发布消息。
	PUBLISH

	// PUBACK 客户端到服务器，或服务器到客户端。发布 QoS 1 消息的确认。
	PUBACK

	// PUBREC 客户端到服务器，或服务器到客户端。发布收到的 QoS 2 消息。保证交付第 1 部分。
	PUBREC

	// PUBREL 客户端到服务器，或服务器到客户端。发布 QoS 2 消息的版本。保证交付第 1 部分。
	PUBREL

	// PUBCOMP 客户端到服务器，或服务器到客户端。发布完整的 QoS 2 消息。保证交付第 3 部分。
	PUBCOMP

	// SUBSCRIBE 订阅客户端到服务器。客户端订阅请求。
	SUBSCRIBE

	// SUBACK 服务器到客户端。订阅确认。
	SUBACK

	// UNSUBSCRIBE 取消订阅客户端到服务器。取消订阅请求。
	UNSUBSCRIBE

	// UNSUBACK 服务器到客户端。退订确认。
	UNSUBACK

	// PINGREQ 客户端到服务器。 PING 请求。
	PINGREQ

	// PINGRESP 服务器到客户端。平响应。
	PINGRESP

	// DISCONNECT 断开客户端到服务器的连接。客户端正在断开连接。
	DISCONNECT

	// AUTH 是保留值，应视为无效消息类型。
	AUTH
)

var typeDescription = [AUTH + 1]string{
	"Reserved",
	"Client request to connect to Server",
	"Accept acknowledgement",
	"Publish message",
	"Publish acknowledgement",
	"Publish received (assured delivery part 1)",
	"Publish release (assured delivery part 2)",
	"Publish complete (assured delivery part 3)",
	"Client subscribe request",
	"Subscribe acknowledgement",
	"Unsubscribe request",
	"Unsubscribe acknowledgement",
	"PING request",
	"PING response",
	"Client is disconnecting",
	"Auth",
}

type Conn struct {
	conn   net.Conn
	maxBuf int
	close  sync.Once
	exit   chan struct{}
}

func newConn(c net.Conn) *Conn {
	return &Conn{
		maxBuf: 10240,
		conn:   c,
		exit:   make(chan struct{}, 1), // 设置为1、防止阻塞
	}
}

// 读取消息
func (c *Conn) Read() {
	go func() {
		defer func() {
			log.Println("exit-read")
		}()
		for {
			select {
			case <-c.exit:
				log.Println("关闭读协程")
				return
			default:
				buf := make([]byte, c.maxBuf)
				i, err := c.conn.Read(buf)
				if err != nil && err.Error() != "EOF" {
					log.Println("err", err)
					c.Close()
					return
				}
				if i > 0 {
					c.Decode(buf[:i])
				}
			}
		}
	}()

}

// 向连接中写入数据包
func (c *Conn) Write(buf []byte) {
	write, err := c.conn.Write(buf)
	if err != nil {
		log.Println("err", err)
		return
	}
	log.Println("成功写入数据包大小:", write)
}

func (c *Conn) Auth() {
	// 返回连接成功消息
	c.Write([]byte{0x20, 0x02, 0x00, 0x00})

}

// Decode 解析数据包
func (c *Conn) Decode(buf []byte) {
	if len(buf) < 1 {
		log.Println("err-length")
		return
	}
	mType := buf[0] >> 0x04
	c.parseType(mType)
}

// Close 关闭服务
func (c *Conn) Close() {
	c.close.Do(func() {
		log.Println("关闭服务连接")
		err := c.conn.Close()
		if err != nil {
			log.Println("err", err)
			return
		}
		c.exit <- struct{}{}
	})

}

func (c *Conn) parseType(t byte) {
	switch t {
	case CONNECT:
		log.Println(typeDescription[t])
		c.Auth()
	case CONNACK:
		log.Println(typeDescription[t])
	case PUBLISH:
		log.Println(typeDescription[t])
	case PUBACK:
		log.Println(typeDescription[t])
	case PUBREC:
		log.Println(typeDescription[t])
	case PUBREL:
		log.Println(typeDescription[t])
	case PUBCOMP:
		log.Println(typeDescription[t])
	case SUBSCRIBE:
		log.Println(typeDescription[t])
	case SUBACK:
		log.Println(typeDescription[t])
	case UNSUBACK:
		log.Println(typeDescription[t])
	case UNSUBSCRIBE:
		log.Println(typeDescription[t])
	case PINGREQ:
		log.Println(typeDescription[t])
	case PINGRESP:
		log.Println(typeDescription[t])
	case DISCONNECT:
		log.Println(typeDescription[t])
		c.Close()
	case AUTH:
		log.Println(typeDescription[t])
	case RESERVED:
		log.Println(typeDescription[t])
	default:
		log.Println("err-type")
	}
	return
}

// 开启监听
func listen(l net.Listener) {
	go func() {
		defer func() {
			log.Println("exit-connListen")
		}()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("err-conn", err)
				return
			}
			log.Println("本次链接地址:", conn.RemoteAddr().String())
			c := newConn(conn)
			c.Read()
		}
	}()
}

func main() {
	log.Println("Listing in port 1883")
	defer func() {
		log.Println("exit-main")
		if r := recover(); r != nil {
			log.Println("err", r)
		}
	}()
	lister, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Println("lister-error", err)
		return
	}
	listen(lister)
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-exitChan
	log.Println("service received signal:", sig.String())
}
