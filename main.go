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
	// 在开发模式下，Wails 会自动使用 wails.json 中配置的 devServer
	// 如果开发服务器可用，Wails 会优先使用它而不是嵌入的静态文件
	var assetServer *assetserver.Options

	if isDev {
		logger.Info("开发模式：将使用前端开发服务器（devServer）")
		logger.Info("确保前端开发服务器正在运行: http://localhost:34115")
		// 在开发模式下，仍然设置 AssetServer（作为后备）
		// Wails 会检查 devServer 是否可用，如果可用则使用 devServer
		assetServer = &assetserver.Options{
			Assets: assets,
		}
	} else {
		// 生产模式：使用嵌入的静态文件
		assetServer = &assetserver.Options{
			Assets: assets,
		}
		logger.Info("生产模式：使用嵌入的静态文件")
	}

	err := wails.Run(&options.App{
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
	})

	if err != nil {
		logger.Fatalf("应用启动失败: %v", err)
	}
}
