package app

import (
	"app-frame-work/config"
	"fmt"
	"testing"
)

func TestBuildFrameAppContext(t *testing.T) {
	app := BuildFrameAppContext()
	configData := app.Config
	fmt.Println(configData)
}

func TestBuildFrameAppContextWithConfig(t *testing.T) {
	app := BuildFrameAppContextWithConfig(config.NewDefaultAppConfig())
	configData := app.Config
	fmt.Println(configData)
}

func TestStart(t *testing.T) {
	app := BuildFrameAppContext()
	err := app.Start(&app)
	if err != nil {
		return
	}
}
