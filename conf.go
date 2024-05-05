package conf

import (
	"reflect"
	"strings"

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
		p = New(name + "-" + o.PrimaryAliss)
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
	return
}

func (c *Confhub) Exist() bool {
	return zfile.FileExist(c.fullpath)
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
	case reflect.Struct:
		c.Core.SetDefault(key, ztype.ToMap(value))
		return
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

func (c *Confhub) ConfigChange(fn func(e fsnotify.Event)) {
	c.Core.WatchConfig()
	c.Core.OnConfigChange(func(in fsnotify.Event) {
		_ = c.GetAll(true)
		fn(in)
	})
}

func (c *Confhub) GetAll(force ...bool) ztype.Map {
	if c.data == nil || (len(force) > 0 && force[0]) {
		c.data = c.Core.AllSettings()
	}
	return c.data
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
