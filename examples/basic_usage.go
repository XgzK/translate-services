//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"untitled/internal/translator/deeplx"
)

// printSeparator 打印分隔线，参数: 无，返回: 无
func printSeparator() {
	fmt.Println(strings.Repeat("=", 70))
}

// printResult 打印翻译结果，参数: 名称与结果，返回: 无
func printResult(name string, result *deeplx.TranslationResult) {
	fmt.Printf("【%s】\n", name)
	if result.Success {
		fmt.Printf("  ✅ 成功！\n")
		fmt.Printf("  译文: %s\n", result.TranslatedText)
		fmt.Printf("  源语言: %s → 目标语言: %s\n", result.SourceLang, result.TargetLang)
		if result.RawResponse != nil {
			fmt.Printf("  响应代码: %d\n", result.RawResponse.Code)
		}
	} else {
		fmt.Printf("  ❌ 失败！\n")
		fmt.Printf("  错误: %s\n", result.ErrorMessage)
	}
	fmt.Println(strings.Repeat("-", 70))
}

// main 示例主函数，参数: 无，返回: 无
func main() {
	// 从环境变量获取 API 密钥（推荐方式，避免硬编码喵～）
	apiKey := "sk-jotjCcLK2bhfbIvMsgDSMhvgFRXviVWiDaC4af4400LIab8V" //os.Getenv("DEEPLX_API_KEY")
	//if apiKey == "" {
	//	fmt.Println("❌ 错误：请设置环境变量 DEEPLX_API_KEY")
	//	fmt.Println()
	//	fmt.Println("设置方式：")
	//	fmt.Println("  Windows (PowerShell): $env:DEEPLX_API_KEY=\"sk-your-key\"")
	//	fmt.Println("  Windows (CMD):        set DEEPLX_API_KEY=sk-your-key")
	//	fmt.Println("  Linux/Mac:            export DEEPLX_API_KEY=\"sk-your-key\"")
	//	fmt.Println()
	//	fmt.Println("或者直接在代码中设置（不推荐）：")
	//	fmt.Println("  apiKey := \"sk-your-api-key\"")
	//	os.Exit(1)
	//}

	// 创建翻译器实例
	translator, err := deeplx.NewTranslator(apiKey)
	if err != nil {
		fmt.Printf("❌ 初始化翻译器失败: %v\n", err)
		os.Exit(1)
	}

	printSeparator()
	fmt.Println("🐱 DeepLX API 使用示例 (浮浮酱为您演示喵～)")
	printSeparator()
	fmt.Println()

	// 示例 1: 基本翻译（指定源语言）
	result1 := translator.Translate("Hello, world!", "ZH", "EN")
	printResult("示例 1: 英译中", result1)
	fmt.Println()

	//// 示例 2: 自动检测源语言
	//result2 := translator.Translate("你好，世界！", "EN")
	//printResult("示例 2: 自动检测源语言（中译英）", result2)
	//fmt.Println()
	//
	//// 示例 3: 长文本翻译
	//longText := "Machine learning is a subset of artificial intelligence that " +
	//	"provides systems the ability to automatically learn and improve " +
	//	"from experience without being explicitly programmed."
	//result3 := translator.Translate(longText, "ZH", "EN")
	//printResult("示例 3: 长文本翻译", result3)
	//fmt.Println()
	//
	//// 示例 4: 多语言翻译（法译日）
	//result4 := translator.Translate("Bonjour, comment allez-vous?", "JA", "FR")
	//printResult("示例 4: 多语言翻译（法译日）", result4)
	//fmt.Println()
	//
	//// 示例 5: 使用指定模型翻译（如果支持）
	//result5 := translator.TranslateWithModel(
	//	"Artificial Intelligence is transforming our world.",
	//	"ZH",
	//	"gpt-4", // 模型名称
	//	"EN",
	//)
	//printResult("示例 5: 使用指定模型翻译", result5)
	//fmt.Println()

	printSeparator()
	fmt.Println("✨ 示例运行完成！(浮浮酱完成任务了呢) o(*￣︶￣*)o")
	printSeparator()
}
