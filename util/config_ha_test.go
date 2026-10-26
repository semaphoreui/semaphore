package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHAConfig_EngineType(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *HAConfig
		expected HAEngineType
	}{
		{"nil config", nil, HAEngineRedis},
		{"empty engine", &HAConfig{}, HAEngineRedis},
		{"redis", &HAConfig{Engine: HAEngineRedis}, HAEngineRedis},
		{"redis_sentinel", &HAConfig{Engine: HAEngineRedisSentinel}, HAEngineRedisSentinel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cfg.EngineType())
		})
	}
}

func TestHARedisSentinelConfig_IsSet(t *testing.T) {
	var nilCfg *HARedisSentinelConfig
	assert.False(t, nilCfg.IsSet())
	assert.False(t, (&HARedisSentinelConfig{}).IsSet())
	assert.False(t, (&HARedisSentinelConfig{Pass: "x"}).IsSet(), "credentials alone do not select Sentinel")
	assert.True(t, (&HARedisSentinelConfig{MasterName: "mymaster"}).IsSet())
	assert.True(t, (&HARedisSentinelConfig{Addrs: []string{"s1:26379"}}).IsSet())
}

func TestHAConfig_JSON(t *testing.T) {
	raw := `{
		"enabled": true,
		"engine": "redis_sentinel",
		"redis_sentinel": {
			"master_name": "mymaster",
			"addrs": ["s1:26379", "s2:26379"],
			"sentinel_user": "sentinel",
			"sentinel_pass": "s3cret",
			"db": 2,
			"user": "app",
			"pass": "p4ss",
			"tls": true,
			"tls_skip_verify": true
		}
	}`
	var cfg HAConfig
	require.NoError(t, json.Unmarshal([]byte(raw), &cfg))

	assert.Equal(t, HAEngineRedisSentinel, cfg.EngineType())
	require.NotNil(t, cfg.RedisSentinel)
	assert.Equal(t, "mymaster", cfg.RedisSentinel.MasterName)
	assert.Equal(t, []string{"s1:26379", "s2:26379"}, cfg.RedisSentinel.Addrs)
	assert.Equal(t, "sentinel", cfg.RedisSentinel.SentinelUser)
	assert.Equal(t, "s3cret", cfg.RedisSentinel.SentinelPass)
	assert.Equal(t, 2, cfg.RedisSentinel.DB)
	assert.Equal(t, "app", cfg.RedisSentinel.User)
	assert.Equal(t, "p4ss", cfg.RedisSentinel.Pass)
	assert.True(t, cfg.RedisSentinel.TLS)
	assert.True(t, cfg.RedisSentinel.TLSSkipVerify)
	assert.Nil(t, cfg.Redis)
}

func TestLoadEnvironmentToObject_HASentinel(t *testing.T) {
	t.Setenv("SEMAPHORE_HA_ENABLED", "true")
	t.Setenv("SEMAPHORE_HA_ENGINE", "redis_sentinel")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_MASTER_NAME", "mymaster")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_ADDRS", `["s1:26379","s2:26379"]`)
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_SENTINEL_USER", "sentinel")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_SENTINEL_PASS", "s3cret")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_DB", "3")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_USER", "app")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_PASS", "p4ss")
	t.Setenv("SEMAPHORE_HA_REDIS_SENTINEL_TLS", "true")

	cfg := NewConfigType()
	sensitive, err := loadEnvironmentToObject(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.HA)
	assert.True(t, cfg.HA.Enabled)
	assert.Equal(t, HAEngineRedisSentinel, cfg.HA.EngineType())
	require.NotNil(t, cfg.HA.RedisSentinel)
	assert.Equal(t, "mymaster", cfg.HA.RedisSentinel.MasterName)
	assert.Equal(t, []string{"s1:26379", "s2:26379"}, cfg.HA.RedisSentinel.Addrs)
	assert.Equal(t, "sentinel", cfg.HA.RedisSentinel.SentinelUser)
	assert.Equal(t, "s3cret", cfg.HA.RedisSentinel.SentinelPass)
	assert.Equal(t, 3, cfg.HA.RedisSentinel.DB)
	assert.Equal(t, "app", cfg.HA.RedisSentinel.User)
	assert.Equal(t, "p4ss", cfg.HA.RedisSentinel.Pass)
	assert.True(t, cfg.HA.RedisSentinel.TLS)
	assert.False(t, cfg.HA.RedisSentinel.TLSSkipVerify)

	assert.Contains(t, sensitive, "SEMAPHORE_HA_REDIS_SENTINEL_PASS")
	assert.Contains(t, sensitive, "SEMAPHORE_HA_REDIS_SENTINEL_SENTINEL_PASS")
}

func TestLoadConfigDefaults_HAEngine(t *testing.T) {
	Config = NewConfigType()
	Config.HA = &HAConfig{Enabled: true}

	loadConfigDefaults()

	assert.Equal(t, HAEngineRedis, Config.HA.Engine, "the default engine is applied to a present ha block")
}

func TestValidateHAConfig(t *testing.T) {
	sentinel := &HARedisSentinelConfig{MasterName: "mymaster", Addrs: []string{"s1:26379"}}

	tests := []struct {
		name      string
		cfg       *HAConfig
		wantErr   string
		wantWarns int
	}{
		{"nil block", nil, "", 0},
		{"disabled block is not checked", &HAConfig{Enabled: false, Engine: HAEngineRedisSentinel}, "", 0},
		{"redis by default", &HAConfig{Enabled: true, Redis: &HARedisConfig{Addr: "r:6379"}}, "", 0},
		{"redis with empty blocks (env loader allocates them)", &HAConfig{Enabled: true, Redis: &HARedisConfig{}, RedisSentinel: &HARedisSentinelConfig{}}, "", 0},
		{"redis explicit", &HAConfig{Enabled: true, Engine: HAEngineRedis}, "", 0},
		{"sentinel valid", &HAConfig{Enabled: true, Engine: HAEngineRedisSentinel, RedisSentinel: sentinel}, "", 0},
		{"sentinel without block", &HAConfig{Enabled: true, Engine: HAEngineRedisSentinel}, "master_name is required", 0},
		{"sentinel without master", &HAConfig{Enabled: true, Engine: HAEngineRedisSentinel, RedisSentinel: &HARedisSentinelConfig{Addrs: []string{"s1:26379"}}}, "master_name is required", 0},
		{"sentinel without addrs", &HAConfig{Enabled: true, Engine: HAEngineRedisSentinel, RedisSentinel: &HARedisSentinelConfig{MasterName: "m"}}, "at least one sentinel", 0},
		{"unknown engine", &HAConfig{Enabled: true, Engine: "etcd"}, "Engine", 0},
		{"unknown engine on disabled block still rejected", &HAConfig{Enabled: false, Engine: "etcd"}, "Engine", 0},
		{"sentinel ignores redis.addr with a warning", &HAConfig{Enabled: true, Engine: HAEngineRedisSentinel, RedisSentinel: sentinel, Redis: &HARedisConfig{Addr: "r:6379"}}, "", 1},
		{"redis ignores sentinel block with a warning", &HAConfig{Enabled: true, Redis: &HARedisConfig{Addr: "r:6379"}, RedisSentinel: sentinel}, "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := validateHAConfig(tt.cfg)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, warnings, tt.wantWarns)
		})
	}
}

func TestValidateConfig_HA(t *testing.T) {
	Config = new(ConfigType)
	Config.Port = ":3000"
	Config.Dialect = DbDriverSQLite
	Config.CookieHash = "0Sn+edH3doJ4EO4Rl49Y0KrxjUkXuVtR5zKHGGWerxQ="
	Config.CookieEncryption = Config.CookieHash
	Config.AccessKeyEncryption = Config.CookieHash
	Config.GitClientId = GoGitClientId
	Config.GitSubmoduleJobs = 4
	Config.EmailTlsMinVersion = "1.2"

	Config.HA = &HAConfig{Enabled: true, Redis: &HARedisConfig{Addr: "r:6379"}}
	validateConfig()
	assert.Empty(t, HAWarnings())

	Config.HA = &HAConfig{Enabled: true, Engine: HAEngineRedisSentinel}
	assert.Panics(t, func() { validateConfig() })

	Config.HA = &HAConfig{
		Enabled:       true,
		Redis:         &HARedisConfig{Addr: "r:6379"},
		RedisSentinel: &HARedisSentinelConfig{MasterName: "m", Addrs: []string{"s1:26379"}},
	}
	validateConfig()
	assert.Len(t, HAWarnings(), 1)
}
