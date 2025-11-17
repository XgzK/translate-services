//go:build ignore

package main

import (
	"context"
	"fmt"
	"strings"

	"untitled/internal/translation"
	"untitled/internal/translator/deeplx"
)

// printResponse 打印翻译响应，参数: 标题、响应、错误，返回: 无
func printResponse(title string, resp interface{}, err error) {
	fmt.Printf("\n【%s】\n", title)
	fmt.Println(strings.Repeat("-", 70))

	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	// 类型断言，获取谷歌格式响应
	if googleResp, ok := resp.(*translation.Response); ok {
		fmt.Printf("✅ 成功！\n")
		fmt.Printf("源语言: %s\n", googleResp.Src)

		if len(googleResp.Sentences) > 0 {
			fmt.Printf("原文: %s\n", googleResp.Sentences[0].Orig)
			fmt.Printf("译文: %s\n", googleResp.Sentences[0].Trans)
		}

		if googleResp.LDResult != nil {
			if len(googleResp.LDResult.Srclangs) > 0 {
				fmt.Printf("检测到的语言: %v\n", googleResp.LDResult.Srclangs)
			}
			if len(googleResp.LDResult.SrclangsConfidences) > 0 {
				fmt.Printf("置信度: %.2f\n", googleResp.LDResult.SrclangsConfidences[0])
			}
		}
	}

	fmt.Println(strings.Repeat("-", 70))
}

// main 工厂示例入口，参数: 无，返回: 无
func main() {
	// API 密钥（实际使用时应从环境变量或配置文件读取）
	apiKey := "sk-jotjCcLK2bhfbIvMsgDSMhvgFRXviVWiDaC4af4400LIab8V"

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("🐱 翻译服务工厂模式示例 (浮浮酱为您演示喵～)")
	fmt.Println(strings.Repeat("=", 70))

	// ========== 方式一：使用工厂创建服务 ==========
	fmt.Println("\n📦 方式一：使用工厂创建服务")

	factory := deeplx.NewFactory()

	// 查看支持的服务
	fmt.Println("\n支持的翻译服务:")
	for _, serviceType := range factory.GetSupportedServices() {
		info := factory.GetServiceInfo(serviceType)
		fmt.Printf("  • %s: %s\n", serviceType, info)
	}

	// 创建服务配置
	config := &deeplx.TranslationServiceConfig{
		APIKey:  apiKey,
		BaseURL: "", // 使用默认 URL
		Timeout: 30,
	}

	// 创建 DeepLX 服务
	service, err := factory.CreateService(deeplx.ServiceTypeDeepLX, config)
	if err != nil {
		fmt.Printf("创建服务失败: %v\n", err)
		return
	}

	fmt.Printf("\n✅ 成功创建服务: %s\n", service.GetName())
	fmt.Printf("服务可用性: %v\n", service.IsAvailable())

	// 使用服务进行翻译
	resp1, err := service.Translate(
		context.Background(),
		"Hello, world!",
		"EN",
		"ZH",
		[]string{"t"}, // 只请求翻译文本
	)
	printResponse("基本翻译（工厂方式）", resp1, err)

	// ========== 方式二：简化创建 ==========
	fmt.Println("\n📦 方式二：简化创建服务")

	simpleService, err := factory.CreateServiceSimple(deeplx.ServiceTypeDeepLX, apiKey)
	if err != nil {
		fmt.Printf("创建服务失败: %v\n", err)
		return
	}

	resp2, err := simpleService.Translate(
		context.Background(),
		"Good morning!",
		"EN",
		"ZH",
		[]string{"t"},
	)
	printResponse("简化方式翻译", resp2, err)

	// ========== 方式三：请求多种数据 ==========
	fmt.Println("\n📦 方式三：请求多种数据类型")

	resp3, err := service.Translate(
		context.Background(),
		"Hello",
		"EN",
		"ZH",
		[]string{"t", "bd", "rm"}, // 翻译 + 词典 + 音译
	)
	printResponse("多数据类型翻译", resp3, err)

	// ========== 方式四：自动语言检测 ==========
	fmt.Println("\n📦 方式四：自动语言检测")

	resp4, err := service.Translate(
		context.Background(),
		"你好，世界！",
		"auto", // 自动检测源语言
		"EN",
		[]string{"t"},
	)
	printResponse("自动检测语言", resp4, err)

	// ========== 演示多服务切换 ==========
	fmt.Println("\n📦 演示：多服务架构优势")
	fmt.Println("当前可用服务:")

	services := []deeplx.ServiceType{
		deeplx.ServiceTypeDeepLX,
		deeplx.ServiceTypeBaidu,
		deeplx.ServiceTypeYoudao,
	}

	for _, st := range services {
		testService, err := factory.CreateServiceSimple(st, apiKey)
		if err != nil {
			fmt.Printf("  • %s: ❌ 尚未实现 (%v)\n", st, err)
		} else {
			fmt.Printf("  • %s: ✅ 可用\n", testService.GetName())
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✨ 工厂模式的优势 (浮浮酱的设计理念喵～):")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("1. ✅ 统一接口：所有翻译服务实现相同接口")
	fmt.Println("2. ✅ 易于扩展：添加新服务无需修改现有代码 (开放封闭原则)")
	fmt.Println("3. ✅ 灵活切换：可以轻松在不同服务间切换")
	fmt.Println("4. ✅ 配置统一：使用统一的配置结构")
	fmt.Println("5. ✅ 面向接口：依赖抽象而非具体实现 (依赖倒置原则)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\n🐱 浮浮酱用心设计，祝您使用愉快喵～ o(*￣︶￣*)o")
}
