package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("连接到服务器...")
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	fmt.Println("连接成功，开始读取数据...")

	// 使用简单的方式读取数据
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("读取失败:", err)
		return
	}

	fmt.Printf("收到数据 (%d 字节): '%s'\n", n, string(buffer[:n]))

	// 发送用户名
	fmt.Print("输入用户名: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		username := scanner.Text()
		fmt.Printf("发送用户名: %s\n", username)

		_, err = conn.Write([]byte(username + "\n"))
		if err != nil {
			fmt.Println("发送失败:", err)
			return
		}

		// 读取响应
		n, err = conn.Read(buffer)
		if err != nil {
			fmt.Println("读取响应失败:", err)
			return
		}

		fmt.Printf("服务器响应 (%d 字节): '%s'\n", n, string(buffer[:n]))
	}

	fmt.Println("按Enter键退出...")
	scanner.Scan()
}
