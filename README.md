# go_conf

Packaging viper for daily multi-environment configuration

## 快速开始

```go
c := conf.New("app") // 读取 ./app.toml
if err := c.Read(); err != nil {
    log.Fatal(err)
}
fmt.Println(c.Get("server.port").Int())
```

## 环境变量覆盖

开启 `AutomaticEnv` 后，每次 `Get` 都会先尝试环境变量，命中则覆盖配置文件的值。

```go
c := conf.New("app", func(o conf.Options) conf.Options {
    o.AutomaticEnv   = true
    o.EnvPrefix      = "APP" // 环境变量前缀
    o.EnvKeyReplacer = "_"   // 默认就是 "_"，server.port -> APP_SERVER_PORT
    o.AllowEmptyEnv  = true  // 空字符串也视为有效值
    return o
})

c.SetDefault("server.port", 8080)
_ = c.Read()

c.Get("server.port").Int() // APP_SERVER_PORT=9090 时为 9090
```

优先级：`Set` > 环境变量 > 配置文件 > `SetDefault`。

注意：

- 环境变量名不能包含 `.`，`EnvKeyReplacer` 留空时默认用 `_` 替换，嵌套 key `server.port` 对应 `APP_SERVER_PORT`。
- 仅存在于环境变量的 key 可以被 `Get` 读到，但不会出现在 `AllKeys`/`GetAll`/`Unmarshal` 中。
  需要通过 `Unmarshal` 反序列化的字段，请先用 `SetDefault` 或配置文件注册。
- 环境变量值始终是字符串，读取方需要显式转换成目标类型。

## 多环境

`PrimaryAliss` 会额外合并一份同目录、同后缀的 `文件名-别名` 配置：

```go
c := conf.New("app", func(o conf.Options) conf.Options {
    o.PrimaryAliss = "dev" // 合并 app-dev.toml
    return o
})
```

## 热更新

```go
c.ConfigChange(func(e fsnotify.Event) {
    log.Println("config changed:", e.Name)
})
```

viper 自身不是并发安全的，配置热更新期间的高频读取请在调用方自行加锁。
