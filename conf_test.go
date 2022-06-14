package conf_test

import (
	"os"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/zlsgo/conf"
)

func TestDef(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := conf.New("zls")
	defer os.Remove("zls.toml")

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
