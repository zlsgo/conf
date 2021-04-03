package conf_test

import (
	"github.com/sohaha/zlsgo"
	"github.com/zlsgo/conf"
	"os"
	"testing"
)

func TestDef(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := gconf.New("zls")
	c.SetDefault("def", 1)
	c.Set("arr", []struct{ Name string }{{"1"}, {"2"}, {"go"}})
	err := c.Read()
	tt.Equal(true, err == nil)

	c2 := gconf.New("zls.yaml")
	err = c2.Read()
	t.Log(c.GetAll())
	t.Log(c2.GetAll())
	tt.Equal(true, err == nil)
	tt.Equal(c.Core.GetInt("def"), c2.Core.GetInt("def"))

	os.Remove("zls.yaml")
}
