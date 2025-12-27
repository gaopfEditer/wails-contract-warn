package main

import (
	"embed"
	"os"

	"wails-contract-warn/logger"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 初始化日志系统
	// 开发环境使用美化输出，生产环境使用 JSON 格式
	// 可以通过环境变量 LOG_LEVEL 和 LOG_PRETTY 控制
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // 默认 info 级别
	}
	logPretty := os.Getenv("LOG_PRETTY") != "false" // 默认使用美化输出
	logger.Init(logLevel, logPretty)

	logger.Info("应用启动中...")

	// 创建应用实例
	app := NewApp()

	// 检查是否为开发模式
	// Wails 在开发模式下会设置 WAILS_ENV=development
	isDev := os.Getenv("WAILS_ENV") == "development" || os.Getenv("DEV") == "true"

	// 配置 AssetServer
	// 在开发模式下，强制使用开发服务器，不提供嵌入的静态文件
	var assetServer *assetserver.Options

	if isDev {
		logger.Info("开发模式：强制使用前端开发服务器（devServer）")
		logger.Info("确保前端开发服务器正在运行: http://localhost:34115")
		// 在开发模式下，不设置 AssetServer，强制 Wails 使用 devServer
		// 如果开发服务器不可用，Wails 会报错而不是回退到嵌入文件
		assetServer = nil
	} else {
		// 生产模式：使用嵌入的静态文件
		assetServer = &assetserver.Options{
			Assets: assets,
		}
		logger.Info("生产模式：使用嵌入的静态文件")
	}

	appOptions := &options.App{
		Title:            "数据分析",
		Width:            1200,
		Height:           800,
		AssetServer:      assetServer,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	}

	// 在开发模式下，强制使用开发服务器
	if isDev {
		logger.Info("开发模式：AssetServer 已设置为 nil，Wails 将使用 wails.json 中的 devServer")
		logger.Info("如果仍然使用嵌入文件，请检查：")
		logger.Info("1. 开发服务器是否在 http://localhost:34115 运行")
		logger.Info("2. wails.json 中的 devServer 配置是否正确")
		logger.Info("3. 是否使用了 'wails dev' 命令")
	}

	err := wails.Run(appOptions)

	if err != nil {
		logger.Fatalf("应用启动失败: %v", err)
	}
}
