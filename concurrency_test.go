package conf_test

import (
	"testing"

	"github.com/zlsgo/conf"
)

func TestGetAllReturnsIsolatedCopy(t *testing.T) {
	c := conf.New("env-copy")
	c.SetDefault("nest", map[string]interface{}{"a": 1})
	c.SetDefault("list", []interface{}{map[string]interface{}{"name": "x"}})
	_ = c.Read()

	data := c.GetAll(true)
	// 修改返回值不应影响配置本身
	if nest, ok := data["nest"].(map[string]interface{}); ok {
		nest["a"] = 999
	}
	if list, ok := data["list"].([]interface{}); ok {
		if item, ok := list[0].(map[string]interface{}); ok {
			item["name"] = "changed"
		}
	}

	if got := c.Get("nest.a").Int(); got != 1 {
		t.Fatalf("nest.a = %d, want 1 (GetAll must return a copy)", got)
	}
	if got := c.Get("list").Slice().Index(0).Get("name").String(); got != "x" {
		t.Fatalf("list[0].name = %q, want x (GetAll must return a copy)", got)
	}
}
