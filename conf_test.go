package conf_test

import (
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/zlsgo/conf"
)

type appString string
type Info struct {
	Age  int    `z:"a"`
	Name string `z:"name"`
}
type Demo struct {
	Name string    `z:"zls"`
	App  appString `json:"app"`
	Info Info
}

func TestDev(t *testing.T) {
	tt := zlsgo.NewTest(t)

	_ = zfile.WriteFile("test-dev.toml", []byte(`zls = "dev"`))
	defer zfile.Remove("test-dev.toml")

	c := conf.New("test", func(o *conf.Options) {
		o.PrimaryAliss = "dev"
	})

	c.SetDefault("info", map[string]interface{}{"a": "18", "name": "is name"})
	c.SetDefault("info2", Demo{Name: "main2", App: "test2"})
	c.SetDefault("infos", []Demo{{Name: "mains", App: "sss", Info: Info{Age: 18, Name: "is name"}}})
	c.SetDefault("zls", "main")
	c.SetDefault("app", "test")
	tt.NoError(c.Read(), true)

	tt.Equal("dev", c.Get("zls").String())
	tt.Equal("test", c.Get("app").String())

	t.Log(c.GetAll())

	var d Demo
	tt.NoError(c.Unmarshal(&d))
	tt.Equal(c.Get("zls").String(), d.Name)
	tt.Equal(c.Get("app").String(), string(d.App))

	var a appString
	tt.NoError(c.UnmarshalKey("app", &a))
	tt.Equal(c.Get("app").String(), string(a))

	var i *Info
	tt.NoError(c.UnmarshalKey("info", &i))
	tt.Equal(c.Get("info.a").Int(), i.Age)
	tt.Equal(c.Get("info").Get("name").String(), i.Name)

	tt.Equal("test2", c.Get("info2").Get("app").String())
	tt.Equal("18", c.Get("infos").Slice().Index(0).Get("Info").Get("a").String())
}

func TestEnv(t *testing.T) {
	tt := zlsgo.NewTest(t)

	_ = os.Setenv("Z_ZLSGO", "YES")

	c := conf.New("env", func(o *conf.Options) {
		o.AutomaticEnv = true
		o.AutoCreate = true
		o.EnvPrefix = "Z"
	})

	defer zfile.Remove(c.Path())

	c.SetDefault("ZLSGO", "123")
	c.SetDefault("ZLS", "sohaha")

	tt.NoError(c.Read())

	tt.Equal(os.Getenv("Z_ZLSGO"), c.Get("ZLSGO").String())
	tt.EqualTrue(c.Get("ZLSGO").String() != "123")

	tt.Equal("sohaha", c.Get("zls").String())
	tt.Equal("sohaha", c.Get("ZLS").String())

	// defer os.Remove("env.toml")

	tt.Log(c.GetAll())
}

func TestDef(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := conf.New("def", func(o *conf.Options) {
		o.AutoCreate = true
	})
	defer os.Remove(c.Path())

	c.SetDefault("def", 1)
	c.Set("project.name", "DemoApp")
	c.Set("arr", []struct{ Name string }{{"1"}, {"2"}, {"go"}})
	c.Set("arr", []struct{ Name string }{{"1"}, {"2"}, {"go"}})
	t.Log(c.Path())
	c.Read()
	t.Log(c.GetAll())

	c2 := conf.New("def.toml")
	c2.Read()
	t.Log(c2.GetAll())

	tt.Equal(c.Get("project.name").String(), c2.Get("project.name").String())
	tt.Equal(c.Get("def").Int(), c2.Get("def").Int())

	tt.NoError(c.Write())
	tt.Equal(c.Path(), c2.Path())
}

func TestFileName(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := conf.New("zls", func(o *conf.Options) {
		o.AutoCreate = true
		o.FileName = "tmp/data/zls"
	})

	c.SetDefault("test", true)
	defer zfile.Rmdir("tmp/")
	tt.NoError(c.Read())
}

func TestConfigChange(t *testing.T) {
	tt := zlsgo.NewTest(t)
	path := "change.toml"
	zfile.WriteFile(path, []byte(`
[info]
name = 'DemoApp'
`))
	defer zfile.Remove(path)

	c := conf.New("change")

	c.ConfigChange(func(e fsnotify.Event) {
		t.Log(e)
	})
	c.Read()

	tt.Equal("DemoApp", c.GetAll().Get("info.name").String())
	zfile.WriteFile(path, []byte(`
[info]
name = 'NewApp'
key = 'test'
`))

	time.Sleep(time.Millisecond * 500)
	tt.Equal("NewApp", c.GetAll().Get("info.name").String())

	c.Set("info.name", "NewApp2")

	tt.Equal("NewApp2", c.GetAll().Get("info.name").String())
}
