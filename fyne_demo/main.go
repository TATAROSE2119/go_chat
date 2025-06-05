package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// 1. 创建一个新的 Fyne 应用
	myApp := app.New()
	// 2. 新建一个窗口。参数是窗口标题
	myWindow := myApp.NewWindow("Hello, Fyne!")

	// 3. 创建一个标签（Label）和一个按钮（Button）
	label := widget.NewLabel("欢迎使用 Fyne！")
	button := widget.NewButton("点击我", func() {
		// 5. 按钮回调：更新标签文字
		label.SetText("按钮已被点击！")
	})

	// 4. 将标签和按钮放到一个垂直布局容器里
	//    容器底层是一个 VBox：label 在上，button 在下
	content := container.NewVBox(
		label,
		button,
	)

	// 6. 把容器内容设置到窗口中
	myWindow.SetContent(content)
	// 7. 设置窗口初始大小（可选）
	myWindow.Resize(fyne.NewSize(300, 200))
	// 8. 显示窗口并启动应用主循环
	myWindow.ShowAndRun()
}
