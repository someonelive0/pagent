package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"pagent/utils"
)

type MyConfig struct {
	Filename string    `yaml:"filename" json:"filename" xml:"filename,attr"`
	LoadTime time.Time `yaml:"load_time" json:"load_time" xml:"load_time,attr"`

	Version    string `yaml:"version" json:"version"`
	Host       string `yaml:"host" json:"host"`
	ManagePort int    `yaml:"manage_port" json:"manage_port" `

	PcapOutput  PcapOutput  `yaml:"pcap_output" json:"pcap_output"`
	RedisConfig RedisConfig `yaml:"redis" json:"redis"`
}

type PcapOutput struct {
	Device string `yaml:"device" json:"device"`
}

type RedisConfig struct {
	Addrs    []string `yaml:"addrs" json:"addrs"`
	Password string   `yaml:"password" json:"password"`
}

func (p *MyConfig) Dump() []byte {
	b, _ := json.MarshalIndent(p, "", " ")
	return b
}

func LoadConfig(filename string) (*MyConfig, error) {
	// 都配置文件，如果文件不存在则从模块文件tpl复制成配置文件。思路是考虑到不覆盖已有现场配置文件。
	if !utils.ExistedOrCopy(filename, filename+".tpl") {
		return nil, fmt.Errorf("config file [%s] or template file are not found", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("config file [%s] open failed: %s", filename, err)
	}

	myconfig := &MyConfig{
		Filename: filename,
		LoadTime: time.Now(),
	}
	err = yaml.Unmarshal(data, myconfig)
	if err != nil {
		return nil, fmt.Errorf("config file [%s] unmarshal yaml failed: %s", filename, err)
	}

	return myconfig, nil
}
