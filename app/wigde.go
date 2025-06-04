package main

import (
	"bufio"
	"fmt"
	"fyne.io/fyne/v2"
	"net"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type ChatClient struct {
	conn        net.Conn
	chatArea    *widget.RichText
	messageList *container.Scroll
	window      fyne.Window
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("聊天室客户端")
	myWindow.Resize(fyne.NewSize(600, 400))

	client := &ChatClient{window: myWindow}

	// 连接界面
	client.showConnectDialog()

	myWindow.ShowAndRun()
}

func (c *ChatClient) showConnectDialog() {
	serverEntry := widget.NewEntry()
	serverEntry.SetText("localhost:8080")
	serverEntry.SetPlaceHolder("服务器地址:端口")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("用户名")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "服务器地址", Widget: serverEntry},
			{Text: "用户名", Widget: usernameEntry},
		},
	}

	dialog.ShowForm("连接到聊天室", "连接", "取消", form.Items, func(ok bool) {
		if !ok {
			c.window.Close()
			return
		}

		if serverEntry.Text == "" || usernameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("请填写所有字段"), c.window)
			c.showConnectDialog()
			return
		}

		c.connectToServer(serverEntry.Text, usernameEntry.Text)
	}, c.window)
}

func (c *ChatClient) connectToServer(server, username string) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		dialog.ShowError(fmt.Errorf("连接失败: %v", err), c.window)
		c.showConnectDialog()
		return
	}

	c.conn = conn
	c.setupChatInterface(username)
}

func (c *ChatClient) setupChatInterface(username string) {
	// 聊天消息显示区域
	c.chatArea = widget.NewRichText()
	c.chatArea.Wrapping = fyne.TextWrapWord

	scroll := container.NewScroll(c.chatArea)
	scroll.SetMinSize(fyne.NewSize(580, 300))

	// 消息输入区域
	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("输入消息...")
	messageEntry.MultiLine = false

	sendButton := widget.NewButton("发送", func() {
		c.sendMessage(messageEntry.Text)
		messageEntry.SetText("")
	})

	// 回车发送消息
	messageEntry.OnSubmitted = func(text string) {
		c.sendMessage(text)
		messageEntry.SetText("")
	}

	// 底部输入栏
	inputContainer := container.NewBorder(nil, nil, nil, sendButton, messageEntry)

	// 顶部工具栏
	disconnectButton := widget.NewButton("断开连接", func() {
		c.disconnect()
	})

	toolbar := container.NewHBox(
		widget.NewLabel(fmt.Sprintf("用户: %s", username)),
		widget.NewSeparator(),
		disconnectButton,
	)

	// 主布局
	content := container.NewBorder(toolbar, inputContainer, nil, nil, scroll)
	c.window.SetContent(content)

	// 处理用户名认证
	c.authenticateUser(username)

	// 启动消息接收goroutine
	go c.receiveMessages()
}

func (c *ChatClient) authenticateUser(username string) {
	// 读取服务器提示
	buffer := make([]byte, 1024)
	_, err := c.conn.Read(buffer)
	if err != nil {
		c.showError("读取服务器提示失败: " + err.Error())
		return
	}

	// 发送用户名
	_, err = fmt.Fprintf(c.conn, "%s\n", username)
	if err != nil {
		c.showError("发送用户名失败: " + err.Error())
		return
	}

	// 读取认证结果
	n, err := c.conn.Read(buffer)
	if err != nil {
		c.showError("读取认证结果失败: " + err.Error())
		return
	}

	response := strings.TrimSpace(string(buffer[:n]))
	if strings.HasPrefix(response, "ERROR:") {
		c.showError(strings.TrimPrefix(response, "ERROR:"))
		return
	}

	c.addMessage("系统", "✅ 成功连接到聊天室！", "green")
}

func (c *ChatClient) sendMessage(message string) {
	if message == "" {
		return
	}

	if c.conn == nil {
		c.showError("未连接到服务器")
		return
	}

	_, err := fmt.Fprintf(c.conn, "%s\n", message)
	if err != nil {
		c.showError("发送消息失败: " + err.Error())
		return
	}

	if message == "exit" {
		c.disconnect()
	}
}

func (c *ChatClient) receiveMessages() {
	if c.conn == nil {
		return
	}

	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			c.addMessage("系统", "❌ 连接已断开", "red")
			break
		}

		message = strings.TrimSpace(message)
		if message != "" {
			// 解析消息类型并着色
			if strings.Contains(message, "加入了聊天室") {
				c.addMessage("系统", message, "green")
			} else if strings.Contains(message, "离开了聊天室") {
				c.addMessage("系统", message, "orange")
			} else {
				// 普通聊天消息
				parts := strings.SplitN(message, ": ", 2)
				if len(parts) == 2 {
					c.addMessage(parts[0], parts[1], "black")
				} else {
					c.addMessage("", message, "black")
				}
			}
		}
	}
}

func (c *ChatClient) addMessage(sender, message, color string) {
	var displayText string
	if sender != "" {
		displayText = fmt.Sprintf("[%s] %s", sender, message)
	} else {
		displayText = message
	}

	// 在UI线程中更新界面
	currentText := c.chatArea.Text
	newText := currentText + displayText + "\n"
	c.chatArea.SetText(newText)

	// 滚动到底部
	if c.messageList != nil {
		c.messageList.ScrollToBottom()
	}
}

func (c *ChatClient) showError(message string) {
	dialog.ShowError(fmt.Errorf(message), c.window)
}

func (c *ChatClient) disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.showConnectDialog()
}
