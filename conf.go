package conf

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/spf13/viper"
)

type Confhub struct {
	*viper.Viper
	filename   string
	filepath   string
	filesuffix string
	fullpath   string
	Core       *viper.Viper
}

func New(file string, defConfFile ...string) *Confhub {
	var (
		tmp    []string
		suffix string
		tmpLen int
		core   = viper.New()
		name   = file
		path   = "./"
	)
	if strings.Contains(file, "/") {
		tmp := strings.Split(file, "/")
		tmpLen = len(tmp) - 1
		path = strings.Join(tmp[0:tmpLen], "/")
		name = tmp[tmpLen]
	}

	tmp = strings.Split(name, ".")
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
	if len(defConfFile) > 0 {
		def := New(defConfFile[0])
		err := def.Read()
		if err == nil {
			defConf := def.GetAll()
			for k, v := range defConf {
				core.SetDefault(k, v)
			}
		}
	}
	fullpath := (path + name + "." + suffix)
	return &Confhub{
		Viper:      core,
		filename:   name,
		filepath:   path,
		filesuffix: suffix,
		fullpath:   fullpath,
		Core:       core,
	}
}

func (c *Confhub) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return c.Core.Unmarshal(rawVal, opts...)
}

func (c *Confhub) SetDefault(key string, value interface{}) {
	c.Core.SetDefault(key, value)
}

func (c *Confhub) Read() (err error) {
	err = c.Core.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
		data := c.Core.AllKeys()
		if len(data) > 0 {
			err = c.Write()
		}
	}
	return
}

func (c *Confhub) Exist() bool {
	return zfile.FileExist(c.fullpath)
}

func (c *Confhub) Set(key string, value interface{}) {
	c.Core.Set(key, value)
}

func (c *Confhub) Get(key string) (value interface{}) {
	return c.Core.Get(key)
}

func (c *Confhub) ConfigChange(fn func(e fsnotify.Event)) {
	c.Core.WatchConfig()
	c.Core.OnConfigChange(fn)
}

func (c *Confhub) GetAll() map[string]interface{} {
	return c.Core.AllSettings()
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
