package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const FMT string = "%-15s%s\n"

var conn net.Conn

var fileName = make(chan string, 1)

var helps = map[string]string{
	"open": "connect to server.",
	"user": "send username to server.",
	"get":  "download file from server.",
	"put":  "upload file to server.",
	"cd":   "change directory of server.",
	"lcd":  "change directory of local.",
	"pwd":  "print work directory of server.",
	"lpwd": "print work directory of local.",
	"ls":   "list content of  directory of server",
	"lls":  "list content of directory of local",
	"help": "show help.",
	"quit": "exit program.",
}

func ExecCMD(cmd string) {
	checkConn()

	buf := make([]byte, 512)

	log.Println(cmd)

	_, err := conn.Write([]byte(cmd))
	if err != nil {
		log.Println("Write:", err)
		return
	}

	_, err = conn.Read(buf)
	if err != nil {
		log.Println("Read:", err)
		return
	}

	fmt.Printf("%s", buf)

	CodeAnalyze(getResponCode(buf))
}

func Open(addr string) {
	if addr == "" {
		fmt.Println("Usage:open host[:port]")
		return
	}

	if conn != nil {
		fmt.Println("You are conncted to ", addr)
		return
	}

	buf := make([]byte, 512)

	segs := strings.Split(addr, ":")
	//validate ip
	ValidateIP(segs[0])

	if len(segs) > 2 {
		fmt.Println("Usage:open host[:port]")
		return
	}
	//if port == ""
	if len(segs) == 1 {
		addr = addr + ":21"
	}

	du, err := time.ParseDuration("10s")
	if err != nil {
		log.Println("ParseDuration:", err)
		return
	}

	fmt.Println("connecting ", addr)
	conn, err = net.DialTimeout("tcp", addr, du)
	if err != nil {
		log.Println("Dial:", err)
		return
	}

	_, err = conn.Read(buf)
	if err != nil {
		log.Println("Read:", err)
		return
	}

	fmt.Printf("%s", buf)

	CodeAnalyze(getResponCode(buf))
}

func User(user string) {
	if user == "" {
		fmt.Println("Usage:user username")
		return
	}

	cmd := "USER " + user + "\r\n"

	ExecCMD(cmd)
}

func Pass(pass string) {
	cmd := "PASS " + pass + "\r\n"
	ExecCMD(cmd)
}

func List(dir string) {
	fileName <- ""
	Port()

	if dir == "" {
		dir = "/"
	}

	cmd := "LIST " + dir + "\r\n"
	ExecCMD(cmd)
}

func Port() {
	addr := conn.LocalAddr().String()
	segs := strings.Split(addr, ":")
	log.Println("addr:", segs[0])
	n5, n6 := PORT(segs[0])

	cmd := strings.Replace(addr, ",", ".", -1)
	cmd = "PORT " + cmd + "." + strconv.Itoa(n5) + "." + strconv.Itoa(n6) + "\r\n"
	log.Println(cmd)

	ExecCMD(cmd)
}

var listener net.Listener

func PORT(addr string) (n5 int, n6 int) {
	var err error

	rand.Seed(time.Now().Unix())
	port := PORT_MIN + rand.Intn(PORT_RANGE)
	log.Println("port:", port)

	for i := 0; i < PORT_RANGE; i++ {
		listener, err = net.Listen("tcp", addr+":"+strconv.Itoa(port))
		if err != nil {
			log.Println("listen:", port, " ERROR:", err)
			port++
			time.Sleep(1 * time.Millisecond)
			continue
		}
		break
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept:", err)
			return
		}

		defer listener.Close()
		defer conn.Close()

		temp := <-fileName
		if temp == "" {
			buf, err := ioutil.ReadAll(conn)
			if err != nil {
				log.Println("ReadAll:", err)
				return
			}
			fmt.Printf("\n%s\n", buf)
		} else {
			file, err := os.OpenFile(temp, os.O_CREATE, 0666)
			if err != nil {
				log.Println("OpenFIle:", err)
				return
			}
			defer file.Close()

			_, err = io.Copy(file, conn)
			if err != nil {
				log.Println("Copy:", err)
				return
			}
		}
	}()

	return port / 256, port % 256

}

func Lls() {

}

func Lcd(dir string) {
	if dir == "" {
		fmt.Println("Usage: lcd dir")
		return
	}

	err := os.Chdir(dir)
	if err != nil {
		log.Println("Chdir:", err)
		return
	}
}

func Get(path string) {
	if path == "" {
		fmt.Println("Usage: get file")
		return
	}
	fileName <- path
	Port()

	cmd := "RETR " + path + "\r\n"
	ExecCMD(cmd)
}

func Lpwd() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Getwd:", err)
		return
	}
	fmt.Println(dir)
}

func Help(cmd string) {
	if cmd == "" {
		for key, value := range helps {
			fmt.Printf(FMT, key, value)
		}
	} else if helps[cmd] == "" {
		fmt.Println("Unrecognized cmd:", cmd)
		return
	} else {
		fmt.Println(FMT, helps[""])
	}
}

func getResponCode(buf []byte) int {
	temp := string(buf[0:3])
	code, _ := strconv.Atoi(temp)
	return code
}

func checkConn() {
	if conn == nil {
		fmt.Println("Not connected")
		return
	}
}

func ValidateIP(ip string) bool {
	if ip == "" {
		return false
	}

	segs := strings.Split(ip, ".")
	if len(segs) != 4 {
		return false
	}

	for _, v := range segs {
		num, err := strconv.Atoi(v)
		if err != nil {
			return false
		}
		if num < 0 || num > 255 {
			return false
		}
	}
	return true
}
