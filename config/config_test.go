package config

import (
	"testing"
)

func TestNewDefaultAppConfig(t *testing.T) {
	config := NewDefaultAppConfig()
	if config != nil {
		t.Error("默认配置错误")
	}
}
