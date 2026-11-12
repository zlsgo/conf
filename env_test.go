package conf_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/spf13/viper"
	"github.com/zlsgo/conf"
)

// newEnvConf 创建一个开启 AutomaticEnv 的配置实例。
// 配置文件不存在时 Read 返回错误，测试里直接忽略（仅依赖默认值/env）。
func newEnvConf(t *testing.T, prefix string, fn func(o conf.Options) conf.Options) *conf.Confhub {
	t.Helper()
	c := conf.New("env-test", func(o conf.Options) conf.Options {
		o.AutomaticEnv = true
		o.EnvPrefix = prefix
		if fn != nil {
			o = fn(o)
		}
		return o
	})
	if err := c.Read(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("read config: %v", err)
		}
	}
	return c
}

func TestAutomaticEnvOverrideAndFallback(t *testing.T) {
	t.Setenv("AE_TOP", "from-env")
	t.Setenv("AE_NESTED", "nested-env")

	c := newEnvConf(t, "AE", nil)
	c.SetDefault("top", "from-default")
	c.SetDefault("nested", "nested-default")
	c.SetDefault("missing", "missing-default")

	if got := c.Get("top").String(); got != "from-env" {
		t.Fatalf("top = %q, want from-env", got)
	}
	if got := c.Get("nested").String(); got != "nested-env" {
		t.Fatalf("nested = %q, want nested-env", got)
	}
	// 没有对应环境变量时回退默认值
	if got := c.Get("missing").String(); got != "missing-default" {
		t.Fatalf("missing = %q, want missing-default", got)
	}
	// 前缀不匹配时也不应命中
	t.Setenv("OTHER_TOP", "other-env")
	if got := c.Get("top").String(); got != "from-env" {
		t.Fatalf("top = %q, env prefix should not match OTHER", got)
	}
	// 大小写不敏感
	c.SetDefault("MixedCase", "x")
	t.Setenv("AE_MIXEDCASE", "upper")
	if got := c.Get("mixedcase").String(); got != "upper" {
		t.Fatalf("mixedcase = %q, want upper", got)
	}
}

func TestAutomaticEnvDisabledByDefault(t *testing.T) {
	t.Setenv("NOENV_TOP", "from-env")

	c := conf.New("env-test-disabled")
	c.SetDefault("top", "from-default")
	_ = c.Read()

	if got := c.Get("top").String(); got != "from-default" {
		t.Fatalf("top = %q, want from-default (AutomaticEnv is off)", got)
	}
}

func TestAutomaticEnvPriority(t *testing.T) {
	const file = "env-priority.toml"
	_ = zfile.WriteFile(file, []byte(`top = "from-file"`))
	defer zfile.Remove(file)

	t.Setenv("PR_TOP", "from-env")

	c := conf.New("env-priority", func(o conf.Options) conf.Options {
		o.AutomaticEnv = true
		o.EnvPrefix = "PR"
		return o
	})
	c.SetDefault("top", "from-default")
	if err := c.Read(); err != nil {
		t.Fatalf("read config: %v", err)
	}

	// 环境变量 > 配置文件 > 默认值
	if got := c.Get("top").String(); got != "from-env" {
		t.Fatalf("top = %q, want from-env", got)
	}

	// Set 的优先级最高
	c.Set("top", "from-set")
	if got := c.Get("top").String(); got != "from-set" {
		t.Fatalf("top = %q, want from-set", got)
	}
}

func TestAutomaticEnvKeyReplacer(t *testing.T) {
	// 未设置替换符时默认使用 "_"（环境变量名不能包含 "."）
	t.Setenv("DR_DB_HOST", "underscore-host")
	t.Setenv("DR_DB_PORT", "3306")
	def := newEnvConf(t, "DR", nil)
	def.SetDefault("db.host", "default-host")
	def.SetDefault("db.port", 5432)

	if got := def.Get("db.host").String(); got != "underscore-host" {
		t.Fatalf("db.host = %q, want underscore-host", got)
	}
	// 环境变量是字符串，读取方需要显式转换类型
	if got := def.Get("db.port").Int(); got != 3306 {
		t.Fatalf("db.port = %d, want 3306", got)
	}

	// 显式设置替换符时按其替换
	t.Setenv("UR_DB__HOST", "custom-host")
	c := newEnvConf(t, "UR", func(o conf.Options) conf.Options {
		o.EnvKeyReplacer = "__"
		return o
	})
	c.SetDefault("db.host", "default-host")
	if got := c.Get("db.host").String(); got != "custom-host" {
		t.Fatalf("db.host = %q, want custom-host", got)
	}
}

func TestAutomaticEnvAllowEmptyEnv(t *testing.T) {
	// 默认忽略值为空串的环境变量
	t.Setenv("EM_TOP", "")
	ignored := newEnvConf(t, "EM", nil)
	ignored.SetDefault("top", "default")
	if got := ignored.Get("top").String(); got != "default" {
		t.Fatalf("top = %q, want default when empty env is ignored", got)
	}

	// 显式允许空值后，空串会覆盖默认值
	empty := newEnvConf(t, "EM", func(o conf.Options) conf.Options {
		o.AllowEmptyEnv = true
		return o
	})
	empty.SetDefault("top", "default")
	if got := empty.Get("top").String(); got != "" {
		t.Fatalf("top = %q, want empty string", got)
	}
}

func TestAutomaticEnvParentShadowsChild(t *testing.T) {
	// 父级 key 存在环境变量时，会遮蔽子 key（viper 的自动 env 语义）
	t.Setenv("PS_NEST", "parent-env")
	c := newEnvConf(t, "PS", nil)
	c.SetDefault("nest.child", "child-default")

	if got := c.Get("nest.child").String(); got != "" {
		t.Fatalf("nest.child = %q, want empty because parent env shadows it", got)
	}
	if got := c.Get("nest").String(); got != "parent-env" {
		t.Fatalf("nest = %q, want parent-env", got)
	}
}

func TestAutomaticEnvOnlyEnvKeyIsNotEnumerated(t *testing.T) {
	// 仅存在于环境变量的 key 可以被 Get 读到，但不会出现在 AllKeys/GetAll 中
	t.Setenv("ON_ONLY", "only-env")
	c := newEnvConf(t, "ON", nil)
	c.SetDefault("known", "known-default")

	if got := c.Get("only").String(); got != "only-env" {
		t.Fatalf("only = %q, want only-env", got)
	}
	for _, k := range c.AllKeys() {
		if k == "only" {
			t.Fatalf("AllKeys should not contain env-only key %q, got %v", k, c.AllKeys())
		}
	}
	if v, ok := c.GetAll().Get("only").Value().(string); ok && v == "only-env" {
		t.Fatal("GetAll should not contain env-only key")
	}
}

func TestAutomaticEnvUnmarshal(t *testing.T) {
	t.Setenv("UM_DB_HOST", "env-host")
	t.Setenv("UM_ONLY_ENV", "only-env")

	type config struct {
		DB struct {
			Host string
		}
		OnlyEnv string
	}

	c := newEnvConf(t, "UM", func(o conf.Options) conf.Options {
		o.EnvKeyReplacer = "_"
		return o
	})
	// Unmarshal 基于 GetAll，必须先把 key 注册进配置（默认值或配置文件）
	c.SetDefault("db.host", "default-host")

	var v config
	if err := c.Unmarshal(&v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v.DB.Host != "env-host" {
		t.Fatalf("DB.Host = %q, want env-host", v.DB.Host)
	}
	// 没有默认值的纯 env key 不会参与 Unmarshal
	if v.OnlyEnv != "" {
		t.Fatalf("OnlyEnv = %q, want empty (env-only keys are not in GetAll)", v.OnlyEnv)
	}
}

func TestAutomaticEnvChangeAfterRead(t *testing.T) {
	t.Setenv("CH_TOP", "v1")
	c := newEnvConf(t, "CH", nil)
	c.SetDefault("top", "default")

	if got := c.GetAll().Get("top").String(); got != "v1" {
		t.Fatalf("GetAll top = %q, want v1", got)
	}

	t.Setenv("CH_TOP", "v2")
	// Get 实时读取环境变量
	if got := c.Get("top").String(); got != "v2" {
		t.Fatalf("Get top = %q, want v2", got)
	}
	// GetAll 有缓存，需要 force 才刷新
	if got := c.GetAll().Get("top").String(); got != "v1" {
		t.Fatalf("cached GetAll top = %q, want v1", got)
	}
	if got := c.GetAll(true).Get("top").String(); got != "v2" {
		t.Fatalf("forced GetAll top = %q, want v2", got)
	}
}

func TestReadRefreshesGetAllCache(t *testing.T) {
	const file = "env-reload.toml"
	_ = zfile.WriteFile(file, []byte("name = 'first'\n"))
	defer zfile.Remove(file)

	c := conf.New("env-reload")
	if err := c.Read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := c.GetAll().Get("name").String(); got != "first" {
		t.Fatalf("name = %q, want first", got)
	}

	_ = zfile.WriteFile(file, []byte("name = 'second'\n"))
	if err := c.Read(); err != nil {
		t.Fatalf("re-read: %v", err)
	}
	// Read 之后 GetAll 缓存必须失效
	if got := c.GetAll().Get("name").String(); got != "second" {
		t.Fatalf("name = %q, want second after re-read", got)
	}
}

func TestAutomaticEnvPrimaryAlias(t *testing.T) {
	dir := "env-primary"
	_ = zfile.RealPathMkdir(dir)
	mainPath := filepath.Join(dir, "app.toml")
	devPath := filepath.Join(dir, "app-dev.toml")
	_ = zfile.WriteFile(mainPath, []byte("main = 'main'\nshared = 'main-shared'\n"))
	_ = zfile.WriteFile(devPath, []byte("dev = 'dev'\nshared = 'dev-shared'\n"))
	defer zfile.Rmdir(dir)

	t.Setenv("PA_SHARED", "env-shared")

	c := conf.New(filepath.Join(dir, "app"), func(o conf.Options) conf.Options {
		o.PrimaryAliss = "dev"
		o.AutomaticEnv = true
		o.EnvPrefix = "PA"
		return o
	})
	c.SetDefault("shared", "default-shared")
	if err := c.Read(); err != nil {
		t.Fatalf("read: %v", err)
	}

	// primary 文件被合并，环境变量优先级最高
	if got := c.Get("main").String(); got != "main" {
		t.Fatalf("main = %q, want main", got)
	}
	if got := c.Get("dev").String(); got != "dev" {
		t.Fatalf("dev = %q, want dev", got)
	}
	if got := c.Get("shared").String(); got != "env-shared" {
		t.Fatalf("shared = %q, want env-shared", got)
	}
}

func TestPrimaryAliasKeepsSuffix(t *testing.T) {
	dir := "env-primary-suffix"
	_ = zfile.RealPathMkdir(dir)
	mainPath := filepath.Join(dir, "app.json")
	devPath := filepath.Join(dir, "app-dev.json")
	_ = zfile.WriteFile(mainPath, []byte(`{"main":"main"}`))
	_ = zfile.WriteFile(devPath, []byte(`{"dev":"dev"}`))
	defer zfile.Rmdir(dir)

	c := conf.New(filepath.Join(dir, "app.json"), func(o conf.Options) conf.Options {
		o.PrimaryAliss = "dev"
		return o
	})
	if err := c.Read(); err != nil {
		t.Fatalf("read: %v", err)
	}

	if got := c.Get("main").String(); got != "main" {
		t.Fatalf("main = %q, want main", got)
	}
	// primary 必须沿用主配置的后缀，否则会去读 app-dev.toml
	if got := c.Get("dev").String(); got != "dev" {
		t.Fatalf("dev = %q, want dev", got)
	}
}

func TestAutomaticEnvNoPrefix(t *testing.T) {
	// 不设置前缀时，环境变量名就是 key 本身
	t.Setenv("NOPREFIX_KEY", "value")
	c := newEnvConf(t, "", nil)
	c.SetDefault("noprefix_key", "default")

	if got := c.Get("noprefix_key").String(); got != "value" {
		t.Fatalf("noprefix_key = %q, want value", got)
	}
}

func TestAutomaticEnvEmptyReadWithoutFile(t *testing.T) {
	// AutomaticEnv + 无配置文件 + 无默认值时，Read 仍返回文件不存在错误
	c := conf.New("env-no-file-xyz", func(o conf.Options) conf.Options {
		o.AutomaticEnv = true
		o.EnvPrefix = "NF"
		return o
	})
	err := c.Read()
	if err == nil {
		t.Fatal("expected config file not found error")
	}
	var notFound viper.ConfigFileNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}
