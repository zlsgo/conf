package conf

import (
	"reflect"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/spf13/viper"
)

type Confhub struct {
	Core       *viper.Viper
	filename   string
	filepath   string
	filesuffix string
	fullpath   string
	option     Options
	primary    *Confhub
	mu         sync.RWMutex
	data       ztype.Map
}

func New(file string, opt ...func(o Options) Options) *Confhub {
	if file == "" {
		file = "zconf"
	}
	o := Options{
		FileName: file,
	}
	for _, f := range opt {
		o = f(o)
	}

	var (
		tmp    []string
		suffix string
		tmpLen int
		core   = viper.New()
		name   = o.FileName
		path   = "./"
	)

	if o.AutomaticEnv {
		core.AutomaticEnv()
	}

	if o.EnvPrefix != "" {
		core.SetEnvPrefix(o.EnvPrefix)
	}

	if o.EnvKeyReplacer != "" {
		core.SetEnvKeyReplacer(strings.NewReplacer(".", o.EnvKeyReplacer))
	}

	if o.AllowEmptyEnv {
		core.AllowEmptyEnv(true)
	}

	if strings.Contains(name, "/") {
		tmp := strings.Split(name, "/")
		tmpLen = len(tmp) - 1
		path = strings.Join(tmp[0:tmpLen], "/")
		name = tmp[tmpLen]
	}

	tmp = strings.SplitN(name, ".", 2)
	tmpLen = len(tmp) - 1
	if tmpLen >= 1 {
		name = strings.Join(tmp[0:tmpLen], ".")
		suffix = tmp[tmpLen]
	}
	if suffix == "" {
		suffix = "toml"
	}

	path = zfile.RealPath(path, true)
	core.SetConfigName(name)
	core.AddConfigPath(path)

	var p *Confhub
	if o.PrimaryAliss != "" {
		primaryName := path + name + "-" + o.PrimaryAliss + "." + suffix
		primaryOpt := func(cur Options) Options {
			for _, f := range opt {
				cur = f(cur)
			}
			cur.FileName = primaryName
			cur.PrimaryAliss = ""
			return cur
		}
		p = New(primaryName, primaryOpt)
		_ = p.Read()
	}

	return &Confhub{
		primary:    p,
		Core:       core,
		filename:   name,
		filepath:   path,
		filesuffix: suffix,
		fullpath:   path + name + "." + suffix,
		option:     o,
	}
}

func (c *Confhub) Read() (err error) {
	err = c.Core.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			if c.option.AutoCreate || c.primary != nil {
				err = nil
			}
			data := c.Core.AllKeys()

			if !zfile.DirExist(c.filepath) {
				zfile.RealPathMkdir(c.filepath)
			}

			if len(data) > 0 && c.option.AutoCreate {
				err = c.Write()
			}
		}
	}
	if c.primary != nil {
		_ = c.Core.MergeConfigMap(c.primary.GetAll())
	}
	c.mu.Lock()
	c.data = nil
	c.mu.Unlock()
	return
}

func (c *Confhub) Exist() bool {
	return zfile.FileExist(c.fullpath)
}

// flattenDefault converts a map into leaf key/value pairs, e.g.
// key.a, key.b, so that viper falls back to the default for keys
// that are not present in the config file instead of replacing
// the whole map.
func flattenDefault(prefix string, m map[string]interface{}, fn func(key string, value interface{})) {
	for k, v := range m {
		key := prefix + "." + k
		if sm, ok := v.(map[string]interface{}); ok && len(sm) > 0 {
			flattenDefault(key, sm, fn)
			continue
		}
		fn(key, v)
	}
}

func (c *Confhub) SetDefault(key string, value interface{}) {
	vof := reflect.Indirect(zreflect.ValueOf(value))
	switch vof.Kind() {
	case reflect.Slice, reflect.Array:
		switch vof.Type().Elem().Kind() {
		case reflect.Struct:
			c.Core.SetDefault(key, ztype.ToMaps(value))
			return
		}
	case reflect.Map, reflect.Struct:
		var m map[string]interface{}
		if vof.Kind() == reflect.Map {
			var ok bool
			m, ok = value.(map[string]interface{})
			if !ok {
				m = ztype.ToMap(value)
			}
		} else {
			m = ztype.ToMap(value)
		}
		if len(m) > 0 {
			flattenDefault(key, m, c.Core.SetDefault)
			return
		}
	}

	c.Core.SetDefault(key, value)
}

func (c *Confhub) Set(key string, value interface{}) {
	c.Core.Set(key, value)
	_ = c.GetAll(true)
}

func (c *Confhub) Get(key string) (value ztype.Type) {
	return ztype.New(c.Core.Get(key))
}

func (c *Confhub) ConfigChange(fn func(e fsnotify.Event)) (err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	_ = watcher.Close()
	c.Core.WatchConfig()
	c.Core.OnConfigChange(func(in fsnotify.Event) {
		_ = c.GetAll(true)
		fn(in)
	})

	return nil
}

func (c *Confhub) GetAll(force ...bool) ztype.Map {
	c.mu.RLock()
	data := c.data
	c.mu.RUnlock()

	if data == nil || (len(force) > 0 && force[0]) {
		data = ztype.Map(cloneMap(c.Core.AllSettings()))
		c.mu.Lock()
		c.data = data
		c.mu.Unlock()
	}
	return data
}

// cloneMap 深拷贝 map/slice，避免与 viper 内部共享引用。
func cloneMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = cloneValue(v)
	}
	return out
}

func cloneValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return cloneMap(val)
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = cloneValue(item)
		}
		return out
	default:
		return v
	}
}

func (c *Confhub) AllKeys() []string {
	return c.Core.AllKeys()
}

func (c *Confhub) Write(filepath ...string) error {
	f := c.fullpath
	if len(filepath) > 0 {
		f = filepath[0]
	}
	return c.Core.WriteConfigAs(f)
}

func (c *Confhub) Path() string {
	return c.fullpath
}

func (c *Confhub) UnmarshalKey(key string, rawVal interface{}, force ...bool) error {
	return ztype.To(c.GetAll(force...).Get(key).Value(), rawVal)
}

func (c *Confhub) Unmarshal(rawVal interface{}, force ...bool) error {
	return ztype.To(c.GetAll(force...), rawVal)
}
