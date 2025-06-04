package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 配置结构体
type Config struct {
	ServerAddress string `json:"server_address"`
	Username      string `json:"username,omitempty"`
	AutoConnect   bool   `json:"auto_connect"`
}

// 默认配置
var defaultConfig = Config{
	ServerAddress: "localhost:8080",
	AutoConnect:   false,
}

// 读取配置文件
func loadConfig() Config {
	config := defaultConfig

	// 获取配置文件路径（与exe同目录）
	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "client_config.json")

	// 尝试读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 如果配置文件不存在，创建默认配置
		if os.IsNotExist(err) {
			saveConfig(config)
			fmt.Println("📝 已创建默认配置文件: client_config.json")
			fmt.Println("   请编辑此文件设置服务器地址")
			fmt.Println("")
		}
		return config
	}

	// 解析配置文件
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("⚠️  配置文件格式错误: %v\n", err)
		return defaultConfig
	}

	return config
}

// 保存配置文件
func saveConfig(config Config) {
	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "client_config.json")

	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0644)
}

// 显示服务器选择菜单
func selectServer(config *Config) {
	fmt.Println("\n🌐 服务器连接选项:")
	fmt.Println("1. 连接到本地服务器 (localhost:8080)")
	fmt.Printf("2. 连接到配置的服务器 (%s)\n", config.ServerAddress)
	fmt.Println("3. 输入新的服务器地址")
	fmt.Println("4. 退出")

	fmt.Print("\n请选择 (1-4): ")
	stdin := bufio.NewScanner(os.Stdin)
	if !stdin.Scan() {
		return
	}

	choice := strings.TrimSpace(stdin.Text())

	switch choice {
	case "1":
		config.ServerAddress = "localhost:8080"
	case "2":
		// 使用配置文件中的地址
	case "3":
		fmt.Print("请输入服务器地址 (例如: 192.168.1.100:8080): ")
		if stdin.Scan() {
			addr := strings.TrimSpace(stdin.Text())
			if addr != "" {
				config.ServerAddress = addr
				// 询问是否保存
				fmt.Print("是否保存此地址到配置文件? (y/n): ")
				if stdin.Scan() && strings.ToLower(stdin.Text()) == "y" {
					saveConfig(*config)
					fmt.Println("✅ 配置已保存")
				}
			}
		}
	case "4":
		fmt.Println("👋 再见!")
		os.Exit(0)
	default:
		fmt.Println("❌ 无效选择")
		selectServer(config)
		return
	}
}

func main() {
	fmt.Println("🌟 欢迎使用多人聊天室客户端!")

	// 加载配置
	config := loadConfig()

	// 检查命令行参数
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Println("\n使用方法:")
			fmt.Println("  chatclient.exe [服务器地址:端口]")
			fmt.Println("  chatclient.exe                    # 使用菜单选择服务器")
			fmt.Println("  chatclient.exe 192.168.1.100:8080 # 直接连接到指定服务器")
			return
		}
		config.ServerAddress = os.Args[1]
	} else if !config.AutoConnect {
		// 如果没有命令行参数且不是自动连接，显示选择菜单
		selectServer(&config)
	}

	// 连接服务器
	fmt.Printf("\n🏃‍♂️‍➡️ 正在连接到服务器 %s...\n", config.ServerAddress)

	var conn net.Conn
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		conn, err = net.Dial("tcp", config.ServerAddress)
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			fmt.Printf("⚠️  连接失败 (尝试 %d/%d): %v\n", i+1, maxRetries, err)
			fmt.Println("⏳ 2秒后重试...")
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		fmt.Println("❌ 无法连接到服务器:", err)
		fmt.Println("\n请检查:")
		fmt.Println("1. 服务器是否已启动")
		fmt.Println("2. 服务器地址是否正确")
		fmt.Println("3. 防火墙是否阻止了连接")
		fmt.Print("\n按回车键退出...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	defer conn.Close()
	fmt.Println("✅ TCP连接建立成功")

	// 显示连接信息
	fmt.Printf("📍 本地地址: %s\n", conn.LocalAddr())
	fmt.Printf("📍 服务器地址: %s\n", conn.RemoteAddr())

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 读取服务器的用户名提示
	fmt.Println("\n等待服务器响应...")
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("❌ 读取服务器提示失败:", err)
		fmt.Print("\n按回车键退出...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}
	prompt := string(buffer[:n])
	fmt.Printf("🔔 %s", prompt)

	// 输入用户名
	var name string
	if config.Username != "" {
		fmt.Printf("(默认: %s) ", config.Username)
	}

	stdin := bufio.NewScanner(os.Stdin)
	if !stdin.Scan() {
		fmt.Println("❌ 读取用户名失败")
		return
	}

	name = strings.TrimSpace(stdin.Text())
	if name == "" && config.Username != "" {
		name = config.Username
	}

	if name == "" {
		fmt.Println("❌ 用户名不能为空")
		fmt.Print("\n按回车键退出...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	// 发送用户名到服务器
	_, err = fmt.Fprintf(conn, "%s\n", name)
	if err != nil {
		fmt.Println("❌ 发送用户名失败:", err)
		return
	}

	// 读取服务器响应
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("❌ 读取服务器响应失败:", err)
		return
	}

	// 清除读取超时
	conn.SetReadDeadline(time.Time{})

	response := strings.TrimSpace(string(buffer[:n]))

	// 检查服务器响应
	if strings.HasPrefix(response, "ERROR:") {
		fmt.Printf("❌ %s\n", strings.TrimPrefix(response, "ERROR:"))
		fmt.Print("\n按回车键退出...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	} else if strings.HasPrefix(response, "SUCCESS:") {
		fmt.Printf("✅ 欢迎 %s! 成功进入聊天室\n", name)
		fmt.Println("\n📝 使用说明:")
		fmt.Println("  - 直接输入消息并按回车发送")
		fmt.Println("  - 输入 'exit' 退出聊天室")
		fmt.Println("  - 所有消息对聊天室内所有人可见")
		fmt.Println("\n💬 开始聊天吧!\n")
		fmt.Println(strings.Repeat("-", 50))
	} else {
		fmt.Println("⚠️ 收到未知的服务器响应:", response)
		return
	}

	// 启动goroutine来接收服务器消息
	go func() {
		reader := bufio.NewReader(conn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("\n❌ 与服务器的连接已断开:", err)
				break
			}
			message = strings.TrimSpace(message)
			if message != "" {
				// 为不同类型的消息添加不同的格式
				if strings.Contains(message, "加入了聊天室") {
					fmt.Printf("👋 %s\n", message)
				} else if strings.Contains(message, "离开了聊天室") {
					fmt.Printf("👋 %s\n", message)
				} else {
					fmt.Printf("💬 %s\n", message)
				}
			}
		}
		fmt.Println("\n📴 连接已关闭，按回车键退出...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
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
			fmt.Println("\n👋 正在退出聊天室...")
			time.Sleep(1 * time.Second)
			break
		}
	}

	if err := input.Err(); err != nil {
		fmt.Println("❌ 读取输入失败:", err)
	}

	fmt.Println("✅ 已安全退出聊天室")
	fmt.Print("\n按回车键关闭窗口...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
