# config file for pagent
#

version: 3.0

host: 0.0.0.0
manage_port: 9265

# 租户ID，ID必须是整数数字
tenant_id: 5000

# 设置运行的 CPU 数量
cpu_number: 2

# 内部通道缓存大小，默认一百万，占内存8G。根据实际情况扩大到一千万时，大概需要内存32G
channel_size: 10000

# 配置采集来源, 用于 libpcap
capture:
  # 指定本地采集的网卡名称,  all采集所有网卡; 名称使用pagent -l 获取
  devices:
    - \Device\NPF_Loopback

  # 设置网卡采集过滤器，libpcap语法, 支持最大长度200
  filter: tcp and not port 9265 and not port 9266 and not port 22

  # 是否网卡成杂凑模式，默认 false 
  promisc: false

  # pcap抓包参数，一次抓包的数据大小，默认为1518 单位 b
  snaplen: 65536

  # pcap缓存配置，单位MB
  pcap_buffer_size: 100


# zeromq 输出
zeromq:
  addrs:
    - 127.0.0.1:9266
