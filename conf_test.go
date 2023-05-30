package conf_test

import (
	"os"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/zlsgo/conf"
)

func TestDev(t *testing.T) {
	tt := zlsgo.NewTest(t)

	_ = zfile.WriteFile("test-dev.toml", []byte(`zls = "dev"`))
	defer zfile.Remove("test-dev.toml")

	c := conf.New("test", func(o *conf.Option) {
		o.PrimaryAliss = "dev"
	})

	c.SetDefault("zls", "main")
	c.SetDefault("app", "test")
	tt.NoError(c.Read(), true)

	tt.Equal("dev", c.GetString("zls"))
	tt.Equal("test", c.GetString("app"))

	t.Log(c.GetAll())
}

func TestEnv(t *testing.T) {
	tt := zlsgo.NewTest(t)

	_ = os.Setenv("Z_ZLSGO", "YES")

	c := conf.New("env", func(o *conf.Option) {
		o.AutomaticEnv = true
		o.AutoCreate = true
		o.EnvPrefix = "Z"
	})

	defer zfile.Remove(c.Path())

	c.SetDefault("ZLSGO", "123")
	c.SetDefault("ZLS", "sohaha")

	tt.NoError(c.Read())

	tt.Equal(os.Getenv("Z_ZLSGO"), c.GetString("ZLSGO"))
	tt.EqualTrue("123" != c.GetString("ZLSGO"))

	tt.Equal("sohaha", c.GetString("zls"))
	tt.Equal("sohaha", c.GetString("ZLS"))

	// defer os.Remove("env.toml")

	tt.Log(c.GetAll())
}

func TestDef(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := conf.New("zls")

	c.SetDefault("def", 1)
	c.Set("project.name", "DemoApp")
	c.Set("arr", []struct{ Name string }{{"1"}, {"2"}, {"go"}})
	c.Set("arr", []struct{ Name string }{{"1"}, {"2"}, {"go"}})
	t.Log(c.Path())
	tt.NoError(c.Read())
	t.Log(c.GetAll())

	c2 := conf.New("zls.toml")
	tt.NoError(c2.Read())
	t.Log(c2.GetAll())

	tt.Equal(c.GetString("project.name"), c2.GetString("project.name"))
	tt.Equal(c.GetInt("def"), c2.GetInt("def"))

	tt.NoError(c.Write())
	tt.Equal(c.Path(), c2.Path())
}
