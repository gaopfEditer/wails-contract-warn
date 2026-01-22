package main

import (
	"embed"
	"fmt"
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
	// 1. 检查环境变量（Wails 在开发模式下会设置 WAILS_ENV=development）
	// 2. 检查 dist 目录是否有实际构建文件
	logger.Info("检查开发模式...")
	logger.Info("  WAILS_ENV=" + os.Getenv("WAILS_ENV"))
	logger.Info("  DEV=" + os.Getenv("DEV"))
	
	isDev := os.Getenv("WAILS_ENV") == "development" || 
		os.Getenv("DEV") == "true" ||
		os.Getenv("DEV") == "1"
	
	// 如果环境变量未设置，检查 dist 目录是否有实际构建文件
	// 如果 dist 目录不存在或没有 index.html，强制使用开发模式
	if !isDev {
		distIndexPath := "frontend/dist/index.html"
		if _, err := os.Stat(distIndexPath); os.IsNotExist(err) {
			logger.Info("检测到 frontend/dist/index.html 不存在，强制使用开发模式")
			isDev = true
		} else {
			// 检查 dist 目录是否只有 gitkeep 等空文件
			dir, err := os.ReadDir("frontend/dist")
			if err == nil {
				hasRealFiles := false
				for _, entry := range dir {
					// 忽略 .gitkeep 等隐藏文件
					if !entry.IsDir() && entry.Name() != ".gitkeep" && entry.Name() != "gitkeep" {
						hasRealFiles = true
						break
					}
				}
				if !hasRealFiles {
					logger.Info("检测到 frontend/dist 目录没有实际构建文件，强制使用开发模式")
					isDev = true
				}
			}
		}
	}
	
	logger.Info("最终判断: isDev=" + fmt.Sprintf("%v", isDev))

	// 配置 AssetServer
	// 注意：即使设置了 AssetServer，如果 wails.json 中配置了 devServer 且使用 wails dev 启动，
	// Wails 会自动优先使用开发服务器，不会使用嵌入的静态文件
	var assetServer *assetserver.Options

	if isDev {
		logger.Info("开发模式：Wails 将使用 wails.json 中的 devServer")
		logger.Info("确保前端开发服务器正在运行: http://localhost:34115")
		// 在开发模式下，仍然设置 AssetServer（作为后备），但 Wails 会优先使用 devServer
		// 如果开发服务器不可用，Wails 会回退到嵌入文件（但这种情况不应该发生）
		assetServer = &assetserver.Options{
			Assets: assets,
		}
		logger.Info("提示：Wails 会自动检测并使用开发服务器（如果可用）")
		logger.Info("请确保：")
		logger.Info("  1. 前端开发服务器在 http://localhost:34115 运行")
		logger.Info("  2. 在浏览器中访问 http://localhost:34115 能看到应用")
		logger.Info("  3. 使用 'wails dev' 命令启动（不是 'wails build'）")
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

	// 在开发模式下，Wails 会自动使用 wails.json 中的 devServer
	if isDev {
		logger.Info("开发模式：Wails 会自动检测并使用开发服务器")
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
