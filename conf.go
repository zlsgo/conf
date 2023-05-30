package conf

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/spf13/viper"
)

type Confhub struct {
	*viper.Viper
	Core       *viper.Viper
	filename   string
	filepath   string
	filesuffix string
	fullpath   string
	option     Option
	primary    *Confhub
}

func New(file string, opt ...func(o *Option)) *Confhub {
	o := Option{}
	for _, f := range opt {
		f(&o)
	}

	var (
		tmp    []string
		suffix string
		tmpLen int
		core   = viper.New()
		name   = file
		path   = "./"
	)

	if o.AutomaticEnv {
		core.AutomaticEnv()
	}

	if o.EnvPrefix != "" {
		core.SetEnvPrefix(o.EnvPrefix)
	}

	if strings.Contains(file, "/") {
		tmp := strings.Split(file, "/")
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
		Viper:      core,
		Core:       core,
		filename:   name,
		filepath:   path,
		filesuffix: suffix,
		fullpath:   path + name + "." + suffix,
		option:     o,
	}
}

func (c *Confhub) Read() (err error) {
	err = c.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			if c.option.AutoCreate || c.primary != nil {
				err = nil
			}
			data := c.AllKeys()
			if len(data) > 0 && c.option.AutoCreate {
				err = c.Write()
			}
		}
	}
	if c.primary != nil {
		_ = c.MergeConfigMap(c.primary.GetAll())
	}
	return
}

func (c *Confhub) Exist() bool {
	return zfile.FileExist(c.fullpath)
}

func (c *Confhub) Get(key string) (value ztype.Type) {
	return ztype.New(c.Viper.Get(key))
}

func (c *Confhub) ConfigChange(fn func(e fsnotify.Event)) {
	c.WatchConfig()
	c.OnConfigChange(fn)
}

func (c *Confhub) GetAll() map[string]interface{} {
	return c.Viper.AllSettings()
}

func (c *Confhub) Write(filepath ...string) error {
	f := c.fullpath
	if len(filepath) > 0 {
		f = filepath[0]
	}
	return c.Viper.WriteConfigAs(f)
}

func (c *Confhub) Path() string {
	return c.fullpath
}
