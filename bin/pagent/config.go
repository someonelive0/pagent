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

	Version     string `yaml:"version" json:"version"`
	Host        string `yaml:"host" json:"host"`
	ManagePort  int    `yaml:"manage_port" json:"manage_port" `
	TenantId    int    `yaml:"tenant_id" json:"tenant_id" `
	CpuNumber   int    `yaml:"cpu_number" json:"cpu_number"`
	ChannelSize int    `yaml:"channel_size" json:"channel_size"`

	CaptureConfig CaptureConfig `yaml:"capture" json:"capture"`
	ZmqConfig     ZmqConfig     `yaml:"zeromq" json:"zeromq"`
}

type CaptureConfig struct {
	Devices        []string `yaml:"devices" json:"devices"`
	Filter         string   `yaml:"filter" json:"filter"`
	Promisc        bool     `yaml:"promisc" json:"promisc"`
	Snaplen        string   `yaml:"snaplen" json:"snaplen"`
	PcapBufferSize int      `yaml:"pcap_buffer_size" json:"pcap_buffer_size"`
}

type ZmqConfig struct {
	Addrs []string `yaml:"addrs" json:"addrs"`
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
