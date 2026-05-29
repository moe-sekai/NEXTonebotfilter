package filter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/exmeaning/nextonebotfilter/internal/store"
)

func TestParseYAML_OneBotFilter1Compatibility(t *testing.T) {
	// 1. Attempt to read from the actual D:\Python_Project\OneBotFilter-1\config-example.yaml if accessible.
	var yamlData []byte
	var err error

	actualPath := filepath.Join("D:", "Python_Project", "OneBotFilter-1", "config-example.yaml")
	if _, err = os.Stat(actualPath); err == nil {
		yamlData, err = os.ReadFile(actualPath)
		if err != nil {
			t.Logf("Failed to read actual config-example.yaml at %s: %v. Falling back to embedded mock.", actualPath, err)
			yamlData = nil
		}
	}

	// 2. Fallback to mock YAML representation if file is not accessible.
	if len(yamlData) == 0 {
		yamlData = []byte(`# Mock OneBotFilter-1 YAML
server:
  host: "127.0.0.1"
  port: 3939
  suffix: "/ws"
  bot-id: 00000000
  user-agent: "OneBotFilter"

  default:
    user-id:
      mode: "blacklist"
      ids: [ ]
    group-id:
      mode: "whitelist"
      ids: [ ]
  buffer-size: 4096
  sleep-time: 5
  debug: false

bot-apps:
  - name: "bot1"
    uri: "ws://example.com/onebot/v11/ws"
    access-token: "abcd"
    user-id:
      mode: "blacklist"
      ids: [ 222222222, 333333333 ]
    group-id:
      mode: "whitelist"
      ids: [ 12345678 ]
    message:
      mode: "blacklist"
      filters: [ "detail", "b\\d+", "查看 *[\\d.]{2,}", "信息" ]
      prefix: [ "b1", "#" ]
      prefix-replace: "/"

  - name: "bot2"
    uri: "wss://bot2.example.com/onebot/v11/ws"
    group-id:
      mode: "default"
      ids: []
    private-message:
      mode: "whitelist"
      filters: [ "/", "pjsk" ]
      prefix: [ "b2" ]
    group-message:
      mode: "on"
`)
	}

	// 3. Parse the YAML content.
	cfg, err := ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("Failed to parse OneBotFilter-1 YAML: %v", err)
	}

	// 4. Validate Server Config unmarshaling.
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 3939 {
		t.Errorf("Expected port 3939, got %d", cfg.Server.Port)
	}
	if cfg.Server.Suffix != "/ws" {
		t.Errorf("Expected suffix '/ws', got '%s'", cfg.Server.Suffix)
	}
	if string(cfg.Server.BotID) != "00000000" {
		t.Errorf("Expected bot-id '00000000', got '%s'", cfg.Server.BotID)
	}
	if cfg.Server.UserAgent != "OneBotFilter" {
		t.Errorf("Expected user-agent 'OneBotFilter', got '%s'", cfg.Server.UserAgent)
	}
	if cfg.Server.BufferSize != 4096 {
		t.Errorf("Expected buffer-size 4096, got %d", cfg.Server.BufferSize)
	}
	if cfg.Server.SleepTime != 5 {
		t.Errorf("Expected sleep-time 5, got %f", cfg.Server.SleepTime)
	}

	// 5. Validate BotApps Config.
	if len(cfg.BotApps) != 2 {
		t.Fatalf("Expected 2 bot-apps, got %d", len(cfg.BotApps))
	}

	// Check bot1
	bot1 := cfg.BotApps[0]
	if string(bot1.Name) != "bot1" {
		t.Errorf("Expected name 'bot1', got '%s'", bot1.Name)
	}
	if bot1.URI != "ws://example.com/onebot/v11/ws" {
		t.Errorf("Expected uri 'ws://example.com/onebot/v11/ws', got '%s'", bot1.URI)
	}
	if string(bot1.AccessToken) != "abcd" {
		t.Errorf("Expected access-token 'abcd', got '%s'", bot1.AccessToken)
	}
	if bot1.UserID.Mode != "blacklist" || len(bot1.UserID.IDs) != 2 || bot1.UserID.IDs[0] != 222222222 {
		t.Errorf("Expected user-id blacklist with 2 IDs, got %+v", bot1.UserID)
	}
	if bot1.GroupID.Mode != "whitelist" || len(bot1.GroupID.IDs) != 1 || bot1.GroupID.IDs[0] != 12345678 {
		t.Errorf("Expected group-id whitelist with 1 ID, got %+v", bot1.GroupID)
	}
	if bot1.Message.Mode != "blacklist" || len(bot1.Message.Filters) != 4 || bot1.Message.PrefixReplace != "/" {
		t.Errorf("Expected message rules for bot1, got %+v", bot1.Message)
	}

	// Check bot2
	bot2 := cfg.BotApps[1]
	if string(bot2.Name) != "bot2" {
		t.Errorf("Expected name 'bot2', got '%s'", bot2.Name)
	}
	if bot2.URI != "wss://bot2.example.com/onebot/v11/ws" {
		t.Errorf("Expected uri 'wss://bot2.example.com/onebot/v11/ws', got '%s'", bot2.URI)
	}
	if string(bot2.AccessToken) != "" {
		t.Errorf("Expected empty access-token, got '%s'", bot2.AccessToken)
	}
	if bot2.GroupID.Mode != "default" {
		t.Errorf("Expected group-id mode 'default', got '%s'", bot2.GroupID.Mode)
	}
	if bot2.PrivateMessage.Mode != "whitelist" || len(bot2.PrivateMessage.Filters) != 2 || bot2.PrivateMessage.Prefix[0] != "b2" {
		t.Errorf("Expected private-message rules for bot2, got %+v", bot2.PrivateMessage)
	}
	if bot2.GroupMessage.Mode != "on" {
		t.Errorf("Expected group-message mode 'on', got '%s'", bot2.GroupMessage.Mode)
	}

	// 6. Test applying to models.
	gw := &store.FilterGateway{}
	defTpl := &store.FilterTemplate{}
	apps, updatedGw := ApplyYAMLToModels(cfg, gw, defTpl)

	if updatedGw.Host != "127.0.0.1" || updatedGw.Port != 3939 || updatedGw.BotID != "00000000" {
		t.Errorf("Models not updated correctly: host=%s, port=%d, botID=%s", updatedGw.Host, updatedGw.Port, updatedGw.BotID)
	}
	if len(apps) != 2 {
		t.Fatalf("Expected 2 apps from ApplyYAMLToModels, got %d", len(apps))
	}
	if apps[0].Name != "bot1" || apps[1].Name != "bot2" {
		t.Errorf("App names incorrect: app0=%s, app1=%s", apps[0].Name, apps[1].Name)
	}

	// 7. Check exported YAML to ensure it doesn't fail.
	exportedBytes, err := ExportYAML(updatedGw, defTpl, nil, apps)
	if err != nil {
		t.Fatalf("ExportYAML failed: %v", err)
	}
	if len(exportedBytes) == 0 {
		t.Errorf("Exported YAML is empty")
	}
}

func TestFlexibleString_TypeVariations(t *testing.T) {
	yamlInput := []byte(`
server:
  bot-id: 1234567890
bot-apps:
  - name: 888
    uri: "ws://localhost:8080"
    access-token: 99999
`)

	cfg, err := ParseYAML(yamlInput)
	if err != nil {
		t.Fatalf("ParseYAML failed for variations: %v", err)
	}

	if string(cfg.Server.BotID) != "1234567890" {
		t.Errorf("Expected BotID '1234567890', got '%s'", cfg.Server.BotID)
	}
	if len(cfg.BotApps) != 1 {
		t.Fatalf("Expected 1 bot-app, got %d", len(cfg.BotApps))
	}
	if string(cfg.BotApps[0].Name) != "888" {
		t.Errorf("Expected Name '888', got '%s'", cfg.BotApps[0].Name)
	}
	if string(cfg.BotApps[0].AccessToken) != "99999" {
		t.Errorf("Expected AccessToken '99999', got '%s'", cfg.BotApps[0].AccessToken)
	}
}
