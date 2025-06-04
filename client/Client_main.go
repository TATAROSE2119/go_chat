package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	addr := "localhost:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	fmt.Printf("正在连接到服务器 %s...\n", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("❌ 连接服务器失败:", err)
		return
	}
	defer conn.Close()
	fmt.Println("✅ TCP连接建立成功")

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 读取服务器的用户名提示 - 使用简单的字节读取
	fmt.Println("等待服务器提示...")
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("❌ 读取服务器提示失败:", err)
		return
	}
	prompt := string(buffer[:n])
	fmt.Printf("收到服务器提示: %s", prompt)

	// 输入用户名
	stdin := bufio.NewScanner(os.Stdin)
	if !stdin.Scan() {
		fmt.Println("❌ 读取用户名失败")
		return
	}

	name := strings.TrimSpace(stdin.Text())
	if name == "" {
		fmt.Println("❌ 用户名不能为空")
		return
	}

	// 发送用户名到服务器
	fmt.Printf("发送用户名: %s\n", name)
	_, err = fmt.Fprintf(conn, "%s\n", name)
	if err != nil {
		fmt.Println("❌ 发送用户名失败:", err)
		return
	}

	// 读取服务器响应
	fmt.Println("等待服务器响应...")
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("❌ 读取服务器响应失败:", err)
		return
	}

	// 清除读取超时
	conn.SetReadDeadline(time.Time{})

	response := strings.TrimSpace(string(buffer[:n]))
	fmt.Printf("服务器响应: %s\n", response)

	// 检查服务器响应
	if strings.HasPrefix(response, "ERROR:") {
		fmt.Printf("❌ %s\n", strings.TrimPrefix(response, "ERROR:"))
		return
	} else if strings.HasPrefix(response, "SUCCESS:") {
		fmt.Printf("✅ %s\n", strings.TrimPrefix(response, "SUCCESS:"))
		fmt.Println("📝 您可以开始聊天了！输入 'exit' 退出聊天室。")
	} else {
		fmt.Println("⚠️ 收到未知的服务器响应:", response)
		return
	}

	// 启动goroutine来接收服务器消息
	go func() {
		reader := bufio.NewReader(conn) // 为接收消息创建新的reader
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("❌ 读取服务器消息失败:", err)
				break
			}
			message = strings.TrimSpace(message)
			if message != "" {
				fmt.Println(message)
			}
		}
		fmt.Println("📴 与服务器连接已断开。")
		os.Exit(0)
	}()

	// 主循环：读取用户输入并发送到服务器
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		msg := strings.TrimSpace(input.Text())

		// 发送消息到服务器
		_, err := fmt.Fprintf(conn, "%s\n", msg)
		if err != nil {
			fmt.Println("❌ 发送消息失败:", err)
			break
		}

		// 如果用户输入exit，退出程序
		if msg == "exit" {
			fmt.Println("👋 您已退出聊天室。")
			break
		}
	}

	if err := input.Err(); err != nil {
		fmt.Println("❌ 读取输入失败:", err)
	}
}
