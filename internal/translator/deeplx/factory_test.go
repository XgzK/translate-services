package deeplx

import (
	"context"
	"testing"
)

// TestNewFactory 测试工厂创建
// TestNewFactory 测试工厂实例创建，参数: 测试实例，返回: 无
func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Fatal("NewFactory() 返回了 nil")
	}
}

// TestCreateService 测试服务创建
// TestCreateService 测试创建服务逻辑，参数: 测试实例，返回: 无
func TestCreateService(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name        string
		serviceType ServiceType
		config      *TranslationServiceConfig
		wantErr     bool
	}{
		{
			name:        "创建 DeepLX 服务",
			serviceType: ServiceTypeDeepLX,
			config: &TranslationServiceConfig{
				APIKey: "sk-test-key",
			},
			wantErr: false,
		},
		{
			name:        "空配置",
			serviceType: ServiceTypeDeepLX,
			config:      nil,
			wantErr:     true,
		},
		{
			name:        "空 API 密钥",
			serviceType: ServiceTypeDeepLX,
			config: &TranslationServiceConfig{
				APIKey: "",
			},
			wantErr: true,
		},
		{
			name:        "不支持的服务类型",
			serviceType: "unknown",
			config: &TranslationServiceConfig{
				APIKey: "sk-test-key",
			},
			wantErr: true,
		},
		{
			name:        "百度翻译（尚未实现）",
			serviceType: ServiceTypeBaidu,
			config: &TranslationServiceConfig{
				APIKey: "test-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := factory.CreateService(tt.serviceType, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && service == nil {
				t.Error("CreateService() 返回了 nil 服务")
			}

			// 验证服务实现了接口
			if !tt.wantErr && service != nil {
				if !service.IsAvailable() {
					t.Error("服务应该可用")
				}

				name := service.GetName()
				if name == "" {
					t.Error("服务名称为空")
				}
			}
		})
	}
}

// TestCreateServiceSimple 测试简化的服务创建
// TestCreateServiceSimple 测试简化创建方法，参数: 测试实例，返回: 无
func TestCreateServiceSimple(t *testing.T) {
	factory := NewFactory()

	service, err := factory.CreateServiceSimple(ServiceTypeDeepLX, "sk-test-key")
	if err != nil {
		t.Fatalf("CreateServiceSimple() error = %v", err)
	}

	if service == nil {
		t.Fatal("CreateServiceSimple() 返回了 nil 服务")
	}

	if service.GetName() != "DeepLX" {
		t.Errorf("GetName() = %v, want DeepLX", service.GetName())
	}
}

// TestCreateDeepLXServiceWithCustomConfig 测试自定义配置
// TestCreateDeepLXServiceWithCustomConfig 测试自定义配置创建，参数: 测试实例，返回: 无
func TestCreateDeepLXServiceWithCustomConfig(t *testing.T) {
	factory := NewFactory()

	config := &TranslationServiceConfig{
		APIKey:  "sk-test-key",
		BaseURL: "https://custom.example.com/api",
	}

	service, err := factory.CreateService(ServiceTypeDeepLX, config)
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}

	// 验证自定义 URL 已应用
	googleService, ok := service.(*GoogleTranslator)
	if !ok {
		t.Fatal("服务类型不是 *GoogleTranslator")
	}

	if googleService.translator.baseURL != "https://custom.example.com/api" {
		t.Errorf("baseURL = %v, want https://custom.example.com/api",
			googleService.translator.baseURL)
	}
}

// TestGetSupportedServices 测试获取支持的服务列表
// TestGetSupportedServices 测试支持服务列表，参数: 测试实例，返回: 无
func TestGetSupportedServices(t *testing.T) {
	factory := NewFactory()

	services := factory.GetSupportedServices()

	if len(services) == 0 {
		t.Error("支持的服务列表为空")
	}

	// 验证 DeepLX 在列表中
	found := false
	for _, s := range services {
		if s == ServiceTypeDeepLX {
			found = true
			break
		}
	}

	if !found {
		t.Error("支持的服务列表中没有 DeepLX")
	}
}

// TestGetServiceInfo 测试获取服务信息
// TestGetServiceInfo 测试服务信息查询，参数: 测试实例，返回: 无
func TestGetServiceInfo(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name        string
		serviceType ServiceType
		wantEmpty   bool
	}{
		{
			name:        "DeepLX 信息",
			serviceType: ServiceTypeDeepLX,
			wantEmpty:   false,
		},
		{
			name:        "百度翻译信息",
			serviceType: ServiceTypeBaidu,
			wantEmpty:   false,
		},
		{
			name:        "未知服务",
			serviceType: "unknown",
			wantEmpty:   false, // 应该返回 "未知服务类型"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := factory.GetServiceInfo(tt.serviceType)

			if tt.wantEmpty && info != "" {
				t.Errorf("GetServiceInfo() = %v, want empty", info)
			}

			if !tt.wantEmpty && info == "" {
				t.Error("GetServiceInfo() 返回了空字符串")
			}
		})
	}
}

// TestTranslationServiceInterface 测试接口实现
// TestTranslationServiceInterface 测试服务接口实现，参数: 测试实例，返回: 无
func TestTranslationServiceInterface(t *testing.T) {
	factory := NewFactory()

	// 创建服务
	service, err := factory.CreateServiceSimple(ServiceTypeDeepLX, "sk-test-key")
	if err != nil {
		t.Fatalf("CreateServiceSimple() error = %v", err)
	}

	// 验证接口方法
	t.Run("GetName", func(t *testing.T) {
		name := service.GetName()
		if name == "" {
			t.Error("GetName() 返回空字符串")
		}
	})

	t.Run("IsAvailable", func(t *testing.T) {
		available := service.IsAvailable()
		if !available {
			t.Error("IsAvailable() 应该返回 true")
		}
	})

	t.Run("Translate", func(t *testing.T) {
		// 这个测试需要模拟服务器，这里只验证方法存在
		_, err := service.Translate(context.Background(), "test", "en", "zh", []string{"t"})
		// 不验证错误，因为没有真实服务器
		_ = err
	})
}

// ExampleTranslationServiceFactory_CreateService 工厂使用示例
// ExampleTranslationServiceFactory_CreateService 演示创建服务，参数: 无，返回: 无
func ExampleTranslationServiceFactory_CreateService() {
	// 创建工厂
	factory := NewFactory()

	// 配置服务
	config := &TranslationServiceConfig{
		APIKey:  "sk-your-api-key",
		BaseURL: "", // 可选
	}

	// 创建 DeepLX 服务
	service, err := factory.CreateService(ServiceTypeDeepLX, config)
	if err != nil {
		panic(err)
	}

	// 使用服务
	resp, err := service.Translate(context.Background(), "Hello", "en", "zh", []string{"t"})
	if err != nil {
		panic(err)
	}

	_ = resp // 处理响应
}

// ExampleTranslationServiceFactory_CreateServiceSimple 简化使用示例
// ExampleTranslationServiceFactory_CreateServiceSimple 演示简化创建，参数: 无，返回: 无
func ExampleTranslationServiceFactory_CreateServiceSimple() {
	factory := NewFactory()

	// 快速创建服务
	service, err := factory.CreateServiceSimple(ServiceTypeDeepLX, "sk-your-api-key")
	if err != nil {
		panic(err)
	}

	// 检查服务是否可用
	if service.IsAvailable() {
		println("服务可用:", service.GetName())
	}
}
