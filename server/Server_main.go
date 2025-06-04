package main

import (
	"bufio"
	"strings"
)            // Importing bufio for buffered I/O
import "fmt" // Importing fmt for formatted I/O
import "net" // Importing net for networking
// Importing strings for string manipulation
// Importing os for operating system functionality
import "sync" // Importing sync for synchronization primitives

// 创建一个客户端的连接池用于存储所有已连接的客户端连接和用户名

var (
	client = make(map[net.Conn]string) // key: 连接对象, value: 用户名
	mutex  = &sync.Mutex{}             // 使用互斥锁来保护对 client map 的并发访问
)

// 广播函数：将消息发送给所有连接的客户端（除了发送者自己）
func broadcast(sender net.Conn, msg string) {
	mutex.Lock()         // 锁定互斥锁，防止并发访问
	defer mutex.Unlock() // 解锁互斥锁

	for conn := range client {
		if conn != sender { // 如果连接不是发送者的连接
			_, err := fmt.Fprintln(conn, msg) // 向连接发送消息
			if err != nil {                   // 如果发送消息失败
				fmt.Println("发送消息错误:", err)
				mutex.Lock()
				delete(client, conn)
				mutex.Unlock()
			}
		}
	}
}

// 处理客户端连接的函数
func handleConn(conn net.Conn) {
	defer conn.Close() // 确保连接在函数结束时关闭

	// 读取客户端发送的用户名
	reader := bufio.NewReader(conn) // 创建一个新的缓冲读取器
	//让用户输入名称
	fmt.Fprint(conn, "请输入您的用户名: ")       // 提示用户输入用户名
	name, err := reader.ReadString('\n') // 读取用户输入的用户名直到换行符
	if err != nil {
		fmt.Println("读取用户名错误:", err) // 如果读取用户名失败，打印错误
		return                       // 退出处理函数
	}
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Fprint(conn, "用户名不能为空\n")
		return
	}
	mutex.Lock()
	for _, existingName := range client {
		if existingName == name {
			fmt.Fprint(conn, "用户名已存在\n")
			mutex.Unlock()
			return
		}
	}
	client[conn] = name
	mutex.Unlock()
	//向所有人广播新用户加入的消息
	broadcast(conn, fmt.Sprintf("✅ %s 加入了聊天室", name)) // 广播新用户加入的消息
	// 持续读取客户端发送的消息
	for {
		msg, err := reader.ReadString('\n') // 读取客户端发送的消息直到换行符
		if err != nil {
			fmt.Println("读取消息错误:", err) // 如果读取消息失败，打印错误
			break                       // 退出循环，结束处理函数
		}
		msg = strings.TrimSpace(msg) // 去除消息两端的空白字符
		if msg == "exit" {           // 如果收到 "exit" 消息
			fmt.Printf("用户 %s 已退出聊天室\n", name) // 打印用户退出信息
			break                              // 退出循环，结束处理函数
		}
		// 向所有人广播用户发送的消息
		broadcast(conn, fmt.Sprintf(" %s:%s", name, msg))
	}
	//用户断开连接后,清理客户端映射，通知其他人
	mutex.Lock()         // 锁定互斥锁，防止并发访问
	delete(client, conn) // 从 client map 中删除断开连接的客户端
	mutex.Unlock()       // 解锁互斥锁

	broadcast(conn, fmt.Sprintf("❌ %s 离开了聊天室", name)) // 广播用户离开聊天室的消息
	fmt.Printf("用户 %s 已断开连接\n", name)                 // 打印用户断开连接的信息
}

func main() {
	//启动TCP监听 监听8080端口
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err) // 如果监听失败，打印错误并退出
	}
	defer listener.Close() // 确保在程序结束时关闭监听器
	fmt.Println("🚀 聊天服务器已启动，监听端口: 8080")
	//循环接受客户端链接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("连接错误:", err) // 如果接受连接失败，打印错误
			continue                  // 继续等待下一个连接
		}
		// 启动一个新的 goroutine 来处理客户端连接
		go handleConn(conn) // 异步处理连接
	}
}
