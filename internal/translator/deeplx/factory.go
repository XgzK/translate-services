package deeplx

import (
	"fmt"
	"strings"
)

// ServiceType 翻译服务类型 (枚举模式喵～)
type ServiceType string

const (
	ServiceTypeDeepLX ServiceType = "deeplx"  // DeepLX 服务
	ServiceTypeBaidu  ServiceType = "baidu"   // 百度翻译（预留）
	ServiceTypeYoudao ServiceType = "youdao"  // 有道翻译（预留）
	ServiceTypeGoogle ServiceType = "google"  // 谷歌翻译（预留）
	ServiceTypeCustom ServiceType = "custom"  // 自定义服务（预留）
)

// TranslationServiceFactory 翻译服务工厂 (工厂模式：统一创建接口喵～)
type TranslationServiceFactory struct {
	// 可以添加默认配置、缓存等
}

// NewFactory 创建服务工厂实例，参数: 无，返回: TranslationServiceFactory 指针
func NewFactory() *TranslationServiceFactory {
	return &TranslationServiceFactory{}
}

// CreateService 创建指定类型翻译服务，参数: 服务类型与配置，返回: 翻译服务实例或错误
func (f *TranslationServiceFactory) CreateService(
	serviceType ServiceType,
	config *TranslationServiceConfig,
) (TranslationService, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API 密钥不能为空")
	}

	switch strings.ToLower(string(serviceType)) {
	case string(ServiceTypeDeepLX):
		return f.createDeepLXService(config)

	case string(ServiceTypeBaidu):
		// 预留：将来实现百度翻译
		return nil, fmt.Errorf("百度翻译服务尚未实现，敬请期待喵～")

	case string(ServiceTypeYoudao):
		// 预留：将来实现有道翻译
		return nil, fmt.Errorf("有道翻译服务尚未实现，敬请期待喵～")

	case string(ServiceTypeGoogle):
		// 预留：将来实现真实的谷歌翻译
		return nil, fmt.Errorf("谷歌翻译服务尚未实现，敬请期待喵～")

	case string(ServiceTypeCustom):
		// 预留：自定义服务
		return nil, fmt.Errorf("自定义翻译服务需要额外配置，敬请期待喵～")

	default:
		return nil, fmt.Errorf("不支持的服务类型: %s", serviceType)
	}
}

// createDeepLXService 创建 DeepLX 服务，参数: 配置，返回: DeepLX 翻译服务或错误
func (f *TranslationServiceFactory) createDeepLXService(
	config *TranslationServiceConfig,
) (TranslationService, error) {
	// 使用完整配置创建服务（包含 Timeout、BaseURL 等）
	service, err := NewGoogleTranslatorWithConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 DeepLX 服务失败: %w", err)
	}

	return service, nil
}

// CreateServiceSimple 简化创建方法，参数: 服务类型与 APIKey，返回: 翻译服务实例或错误
func (f *TranslationServiceFactory) CreateServiceSimple(
	serviceType ServiceType,
	apiKey string,
) (TranslationService, error) {
	return f.CreateService(serviceType, &TranslationServiceConfig{
		APIKey: apiKey,
	})
}

// GetSupportedServices 获取支持的服务类型，参数: 无，返回: 服务类型切片
func (f *TranslationServiceFactory) GetSupportedServices() []ServiceType {
	return []ServiceType{
		ServiceTypeDeepLX,
		// 以下服务预留，将来可以添加
		// ServiceTypeBaidu,
		// ServiceTypeYoudao,
		// ServiceTypeGoogle,
	}
}

// GetServiceInfo 获取服务描述，参数: 服务类型，返回: 描述字符串
func (f *TranslationServiceFactory) GetServiceInfo(serviceType ServiceType) string {
	info := map[ServiceType]string{
		ServiceTypeDeepLX: "DeepLX - 由 LLM 驱动的高质量翻译服务，兼容 DeepL API",
		ServiceTypeBaidu:  "百度翻译 - 国内主流翻译服务（即将支持）",
		ServiceTypeYoudao: "有道翻译 - 网易旗下翻译服务（即将支持）",
		ServiceTypeGoogle: "谷歌翻译 - Google 官方翻译服务（即将支持）",
		ServiceTypeCustom: "自定义服务 - 支持自定义翻译接口（即将支持）",
	}

	if desc, ok := info[serviceType]; ok {
		return desc
	}

	return "未知服务类型"
}
