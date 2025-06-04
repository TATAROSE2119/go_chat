package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	addr := "localhost:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("❌ 连接服务器失败:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	prompt, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("❌ 读取服务器提示失败:", err)
		return
	}
	fmt.Print(prompt)
	stdin := bufio.NewScanner(os.Stdin)
	if stdin.Scan() {
		name := stdin.Text()
		_, err := fmt.Fprintf(conn, "%s\n", name)
		if err != nil {
			fmt.Println("❌ 发送用户名失败:", err)
			return
		}
	}

	fmt.Println("✅ 连接到服务器成功，您可以开始聊天！")
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("❌ 读取服务器消息失败:", err)
		}
		fmt.Println("📴 与服务器连接已断开。")
		os.Exit(0)
	}()

	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		msg := input.Text()
		_, err := fmt.Fprintf(conn, "%s\n", msg)
		if err != nil {
			fmt.Println("❌ 发送消息失败:", err)
			break
		}
		if msg == "exit" {
			fmt.Println("👋 您已退出聊天室。")
			conn.Close()
			break
		}
	}
}
