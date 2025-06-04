package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// 创建一个客户端的连接池用于存储所有已连接的客户端连接和用户名
var (
	client = make(map[net.Conn]string) // key: 连接对象, value: 用户名
	mutex  = &sync.Mutex{}             // 使用互斥锁来保护对 client map 的并发访问
)

// 获取本机的所有IP地址
func getLocalIPs() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// 跳过未启用的接口
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// 跳过回环接口
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					ips = append(ips, v.IP.String())
				}
			}
		}
	}
	return ips
}

// 广播函数：将消息发送给所有连接的客户端（除了发送者自己）
func broadcast(sender net.Conn, msg string) {
	mutex.Lock()
	defer mutex.Unlock()

	// 创建一个临时切片来存储需要删除的连接
	var toDelete []net.Conn

	for conn := range client {
		if conn != sender {
			_, err := fmt.Fprintln(conn, msg)
			if err != nil {
				fmt.Printf("发送消息到 %s 失败: %v\n", client[conn], err)
				toDelete = append(toDelete, conn)
			}
		}
	}

	// 删除失效的连接
	for _, conn := range toDelete {
		delete(client, conn)
		conn.Close()
	}
}

// 检查用户名是否已存在
func isUserNameExists(name string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	for _, existingName := range client {
		if existingName == name {
			return true
		}
	}
	return false
}

// 添加客户端到连接池
func addClient(conn net.Conn, name string) {
	mutex.Lock()
	defer mutex.Unlock()
	client[conn] = name
}

// 从连接池中移除客户端
func removeClient(conn net.Conn) string {
	mutex.Lock()
	defer mutex.Unlock()

	name := client[conn]
	delete(client, conn)
	return name
}

// 获取当前在线用户数
func getOnlineCount() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(client)
}

// 处理客户端连接的函数
func handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("连接 %s 已关闭\n", conn.RemoteAddr())
	}()

	fmt.Printf("开始处理客户端连接: %s\n", conn.RemoteAddr())

	// 立即发送用户名提示
	fmt.Printf("向 %s 发送用户名提示\n", conn.RemoteAddr())
	prompt := "Enter your username: "
	n, err := conn.Write([]byte(prompt))
	if err != nil {
		fmt.Printf("发送提示失败: %v (写入了 %d 字节)\n", err, n)
		return
	}
	fmt.Printf("成功发送提示 (%d 字节): '%s'\n", n, prompt)

	// 读取客户端发送的用户名
	reader := bufio.NewReader(conn)
	fmt.Printf("等待 %s 输入用户名...\n", conn.RemoteAddr())

	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("读取用户名失败: %v\n", err)
		return
	}

	name = strings.TrimSpace(name)
	fmt.Printf("收到用户名: '%s' (来自 %s)\n", name, conn.RemoteAddr())

	if name == "" {
		fmt.Printf("用户名为空，拒绝连接 %s\n", conn.RemoteAddr())
		conn.Write([]byte("ERROR:Username cannot be empty\n"))
		return
	}

	// 检查用户名是否已存在
	if isUserNameExists(name) {
		fmt.Printf("用户名 '%s' 已存在，拒绝连接 %s\n", name, conn.RemoteAddr())
		conn.Write([]byte("ERROR:Username already exists\n"))
		return
	}

	// 添加客户端到连接池
	addClient(conn, name)

	// 发送成功连接确认
	fmt.Printf("用户 '%s' 连接成功，发送确认\n", name)
	_, err = conn.Write([]byte("SUCCESS:Connected successfully\n"))
	if err != nil {
		fmt.Printf("发送成功确认失败: %v\n", err)
		removeClient(conn)
		return
	}

	// 向所有人广播新用户加入的消息
	onlineCount := getOnlineCount()
	broadcast(conn, fmt.Sprintf("✅ %s 加入了聊天室 (当前在线: %d人)", name, onlineCount))
	fmt.Printf("用户 %s 已加入聊天室 (当前在线: %d人)\n", name, onlineCount)

	// 持续读取客户端发送的消息
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("用户 %s 连接断开: %v\n", name, err)
			break
		}

		msg = strings.TrimSpace(msg)
		if msg == "exit" {
			fmt.Printf("用户 %s 主动退出聊天室\n", name)
			break
		}

		if msg != "" {
			fmt.Printf("收到 %s 的消息: %s\n", name, msg)
			broadcast(conn, fmt.Sprintf("%s: %s", name, msg))
		}
	}

	// 用户断开连接后,清理客户端映射，通知其他人
	userName := removeClient(conn)
	if userName != "" {
		onlineCount := getOnlineCount()
		broadcast(conn, fmt.Sprintf("❌ %s 离开了聊天室 (当前在线: %d人)", userName, onlineCount))
		fmt.Printf("用户 %s 已断开连接 (当前在线: %d人)\n", userName, onlineCount)
	}
}

func main() {
	// 可以通过命令行参数指定端口
	port := ":8080"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	// 启动TCP监听
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(fmt.Sprintf("监听失败: %v", err))
	}
	defer listener.Close()

	fmt.Println("🚀 聊天服务器已启动")
	fmt.Printf("📡 监听端口: %s\n", port)

	// 显示本机IP地址
	fmt.Println("📍 本机IP地址:")
	fmt.Printf("   - localhost%s (本地连接)\n", port)

	ips := getLocalIPs()
	if len(ips) > 0 {
		for _, ip := range ips {
			fmt.Printf("   - %s%s (局域网连接)\n", ip, port)
		}
		fmt.Println("\n💡 提示: 局域网内的其他设备可以使用上述IP地址连接到聊天室")
	} else {
		fmt.Println("   ⚠️  未能获取局域网IP地址")
	}

	fmt.Println("\n等待客户端连接...\n")

	// 循环接受客户端链接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			continue
		}

		fmt.Printf("新客户端连接: %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())
		// 启动一个新的 goroutine 来处理客户端连接
		go handleConn(conn)
	}
}
