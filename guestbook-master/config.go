package main

import (
	"io/ioutil"

	"github.com/silenceper/wechat"
	"github.com/silenceper/wechat/cache"
	"github.com/silenceper/wechat/tcb"
	"gopkg.in/yaml.v2"
)

//Config 配置信息
type Config struct {
	TcbEnv    string `yaml:"tcb_env"`
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

var cfg *Config
var _ = getConfig()

func getConfig() *Config {
	if cfg != nil {
		return cfg
	}
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}
	cfg = &Config{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

var wechatTcb *tcb.Tcb
var _ = getTcb()

func getTcb() *tcb.Tcb {
	if wechatTcb != nil {
		return wechatTcb
	}
	memCache := cache.NewMemory()

	//配置小程序参数
	config := &wechat.Config{
		AppID:     getConfig().AppID,
		AppSecret: getConfig().AppSecret,
		Cache:     memCache,
	}
	wc := wechat.NewWechat(config)
	wechatTcb = wc.GetTcb()
	return wechatTcb
}
