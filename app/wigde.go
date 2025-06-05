package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// 创建新的应用
	myApp := app.New()
	myApp.SetIcon(theme.ComputerIcon()) // 设置应用图标

	// 创建主窗口
	myWindow := myApp.NewWindow("Fyne 简单示例")
	myWindow.Resize(fyne.NewSize(500, 400))
	myWindow.CenterOnScreen()

	// 创建各种UI组件
	createContent(myWindow)

	// 显示窗口并运行应用
	myWindow.ShowAndRun()
}

func createContent(window fyne.Window) {
	// 1. 标题标签
	title := widget.NewLabel("🌟 欢迎使用 Fyne GUI 框架！")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 2. 文本输入框
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("请输入您的姓名...")

	// 3. 多行文本输入
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("在这里输入多行文本...")
	messageEntry.Resize(fyne.NewSize(400, 100))

	// 4. 按钮示例
	greetButton := widget.NewButton("问候", func() {
		name := nameEntry.Text
		if name == "" {
			name = "陌生人"
		}
		greeting := fmt.Sprintf("你好，%s！欢迎使用 Fyne！", name)

		// 显示信息对话框
		dialog.ShowInformation("问候", greeting, window)
	})

	// 5. 复选框
	checkbox := widget.NewCheck("启用高级功能", func(checked bool) {
		if checked {
			fmt.Println("✅ 高级功能已启用")
		} else {
			fmt.Println("❌ 高级功能已禁用")
		}
	})

	// 6. 选择框
	selectWidget := widget.NewSelect([]string{"选项1", "选项2", "选项3"}, func(selected string) {
		fmt.Printf("选择了：%s\n", selected)
	})
	selectWidget.SetSelected("选项1") // 设置默认选择

	// 7. 进度条
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0.0)

	// 8. 进度条控制按钮
	startProgressButton := widget.NewButton("开始进度", func() {
		go func() {
			for i := 0; i <= 100; i++ {
				progressBar.SetValue(float64(i) / 100.0)
				time.Sleep(50 * time.Millisecond)
			}
		}()
	})

	resetProgressButton := widget.NewButton("重置进度", func() {
		progressBar.SetValue(0.0)
	})

	// 9. 滑块
	slider := widget.NewSlider(0, 100)
	slider.Value = 50
	sliderLabel := widget.NewLabel("滑块值: 50")
	slider.OnChanged = func(value float64) {
		sliderLabel.SetText(fmt.Sprintf("滑块值: %.0f", value))
	}

	// 10. 颜色和主题按钮
	themeButton := widget.NewButton("切换主题", func() {
		// 这里演示如何显示确认对话框
		dialog.ShowConfirm("切换主题", "确定要切换到深色主题吗？", func(confirmed bool) {
			if confirmed {
				app.New().Settings().SetTheme(theme.DarkTheme())
				dialog.ShowInformation("主题", "已切换到深色主题（重启生效）", window)
			}
		}, window)
	})

	// 11. 时间显示标签
	timeLabel := widget.NewLabel("")
	updateTime := func() {
		timeLabel.SetText("当前时间: " + time.Now().Format("2006-01-02 15:04:05"))
	}
	updateTime() // 初始更新

	// 定时更新时间
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			updateTime()
		}
	}()

	// 12. 退出按钮
	quitButton := widget.NewButton("退出程序", func() {
		dialog.ShowConfirm("退出", "确定要退出程序吗？", func(confirmed bool) {
			if confirmed {
				window.Close()
			}
		}, window)
	})

	// 创建布局容器
	// 顶部区域
	topContainer := container.NewVBox(
		title,
		widget.NewSeparator(),
	)

	// 输入区域
	inputContainer := container.NewVBox(
		widget.NewLabel("📝 输入演示:"),
		nameEntry,
		messageEntry,
		greetButton,
	)

	// 控件区域
	controlsContainer := container.NewVBox(
		widget.NewLabel("🎛️ 控件演示:"),
		checkbox,
		selectWidget,
		container.NewHBox(sliderLabel, slider),
	)

	// 进度条区域
	progressContainer := container.NewVBox(
		widget.NewLabel("📊 进度条演示:"),
		progressBar,
		container.NewHBox(startProgressButton, resetProgressButton),
	)

	// 底部区域
	bottomContainer := container.NewVBox(
		widget.NewSeparator(),
		timeLabel,
		container.NewHBox(themeButton, quitButton),
	)

	// 使用滚动容器包装所有内容
	content := container.NewVBox(
		topContainer,
		widget.NewSeparator(),
		inputContainer,
		widget.NewSeparator(),
		controlsContainer,
		widget.NewSeparator(),
		progressContainer,
		bottomContainer,
	)

	// 创建滚动容器
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(480, 360))

	// 设置窗口内容
	window.SetContent(scroll)

	// 设置窗口关闭时的确认
	window.SetCloseIntercept(func() {
		dialog.ShowConfirm("退出确认", "确定要关闭应用程序吗？", func(confirmed bool) {
			if confirmed {
				window.Close()
			}
		}, window)
	})
}
