package notifications

import (
	"testing"

	"github.com/semaphoreui/semaphore/util"
)

func TestCreateManager(t *testing.T) {
	// Save original config
	originalConfig := util.Config
	defer func() {
		util.Config = originalConfig
	}()

	// Set up test config
	util.Config = &util.ConfigType{
		Notifications: map[string]util.NotificationServiceConfig{
			"telegram": {
				Enabled: true,
				Token:   "test_token",
				Channel: "test_chat",
			},
		},
	}

	// Test CreateManager
	manager := CreateManager()

	if manager == nil {
		t.Fatal("CreateManager returned nil")
	}

	// Check if telegram service is registered
	service, exists := manager.GetService("telegram")
	if !exists {
		t.Error("Telegram service not registered")
	}

	if service == nil {
		t.Error("Telegram service is nil")
	}

	if service.GetName() != "telegram" {
		t.Errorf("Expected service name 'telegram', got '%s'", service.GetName())
	}

	// Test with legacy config
	util.Config = &util.ConfigType{
		TelegramAlert: true,
		TelegramToken: "legacy_token",
	}

	manager2 := CreateManager()
	service2, exists2 := manager2.GetService("telegram")
	if !exists2 {
		t.Error("Telegram service not registered with legacy config")
	}

	if !service2.IsConfigured() {
		t.Error("Telegram service should be configured with legacy config")
	}
}

func TestManagerDependencyInjection(t *testing.T) {
	// Save original config
	originalConfig := util.Config
	defer func() {
		util.Config = originalConfig
	}()

	// Set up empty test config
	util.Config = &util.ConfigType{}
	
	manager := CreateManager()
	
	// Test that we can get configured services
	services := manager.GetConfiguredServices()
	
	// With no configuration, services should exist but not be configured
	if len(services) != 0 {
		t.Errorf("Expected 0 configured services with empty config, got %d", len(services))
	}
	
	// Test that all expected services are registered
	expectedServices := []string{"telegram", "slack", "gotify", "dingtalk"}
	for _, serviceName := range expectedServices {
		service, exists := manager.GetService(serviceName)
		if !exists {
			t.Errorf("Service '%s' not registered", serviceName)
		}
		if service == nil {
			t.Errorf("Service '%s' is nil", serviceName)
		}
		if service.GetName() != serviceName {
			t.Errorf("Expected service name '%s', got '%s'", serviceName, service.GetName())
		}
	}
}