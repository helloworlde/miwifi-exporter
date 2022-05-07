package collector

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	metrics map[string]*prometheus.Desc
	mutex   sync.Mutex
}

func newGlobalMetric(namespace string, metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, docString, labels, nil)
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		metrics: map[string]*prometheus.Desc{
			"hardware_info": newGlobalMetric(namespace, "hardware_info", "机器信息", []string{"router_mac", "platform", "version", "channel", "sn"}),
			"memory_info":   newGlobalMetric(namespace, "memory_info", "内存信息", []string{"router_mac", "total", "type", "frequency"}),
			"cpu_info":      newGlobalMetric(namespace, "cpu_info", "CPU信息", []string{"router_mac", "core", "frequency"}),
			"uptime":        newGlobalMetric(namespace, "uptime", "启动时间(s)", []string{"router_mac"}),

			"history_device_amount": newGlobalMetric(namespace, "history_device_amount", "历史设备数量", []string{"router_mac"}),
			"online_device_amount":  newGlobalMetric(namespace, "online_device_amount", "在线设备数量", []string{"router_mac"}),

			"cpu_load_percent":     newGlobalMetric(namespace, "cpu_load_percent", "CPU使用率", []string{"router_mac"}),
			"memory_usage_percent": newGlobalMetric(namespace, "memory_usage_percent", "内存使用率", []string{"router_mac"}),
			"memory_usage":         newGlobalMetric(namespace, "memory_usage", "内存使用量", []string{"router_mac"}),

			"temperature": newGlobalMetric(namespace, "temperature", "温度", []string{"router_mac"}),

			"wan_upload_speed":     newGlobalMetric(namespace, "wan_upload_speed", "WAN 当前上传速度(Byte/s)", []string{"router_mac"}),
			"wan_download_speed":   newGlobalMetric(namespace, "wan_download_speed", "WAN 当前下载速度(Byte/s)", []string{"router_mac"}),
			"wan_upload_traffic":   newGlobalMetric(namespace, "wan_upload_traffic", "WAN 上传流量(Byte)", []string{"router_mac"}),
			"wan_download_traffic": newGlobalMetric(namespace, "wan_download_traffic", "WAN 下载流量(Byte)", []string{"router_mac"}),
			"max_download_speed":   newGlobalMetric(namespace, "max_download_speed", "WAN 最大下载速度(Byte/s)", []string{"router_mac"}),
			"max_upload_speed":     newGlobalMetric(namespace, "max_upload_speed", "WAN 最大上传速度(Byte/s)", []string{"router_mac"}),

			"ipv4":      newGlobalMetric(namespace, "ipv4", "路由器外网 IPV4 地址", []string{"router_mac", "ipv4"}),
			"ipv4_mask": newGlobalMetric(namespace, "ipv4_mask", "路由器外网 IPV4 掩码", []string{"router_mac", "ipv4"}),
			"ipv6":      newGlobalMetric(namespace, "ipv6", "路由器外网 IPV6 地址", []string{"router_mac", "ipv6"}),
			"wan_info":  newGlobalMetric(namespace, "wan_info", "WAN 连接信息", []string{"router_mac", "username", "password", "wan_type"}),

			"device_upload_traffic":   newGlobalMetric(namespace, "device_upload_traffic", "设备上传流量(Byte)", []string{"router_mac", "device_mac", "device_name", "ip"}),
			"device_download_traffic": newGlobalMetric(namespace, "device_download_traffic", "设备下载流量(Byte)", []string{"router_mac", "device_mac", "device_name", "ip"}),
			"device_upload_speed":     newGlobalMetric(namespace, "device_upload_speed", "设备上传速度(Byte/s)", []string{"router_mac", "device_mac", "device_name", "ip"}),
			"device_download_speed":   newGlobalMetric(namespace, "device_download_speed", "设备下载速度(Byte/s)", []string{"router_mac", "device_mac", "device_name", "ip"}),
			"device_online_time":      newGlobalMetric(namespace, "device_online_time", "设备在线时间(s)", []string{"router_mac", "device_mac", "device_name", "ip"}),
		},
	}
}

func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // 加锁
	defer c.mutex.Unlock()
	GetMiwifiStatus()
	GetIPtoMAC()
	GetWAN()
	routerMac := DevStatus.Hardware.Mac
	// 硬件信息
	ch <- prometheus.MustNewConstMetric(c.metrics["hardware_info"], prometheus.GaugeValue, 1, routerMac, DevStatus.Hardware.Platform, DevStatus.Hardware.Version, DevStatus.Hardware.Channel, DevStatus.Hardware.Sn)
	// 内存信息
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_info"], prometheus.GaugeValue, 1, routerMac, DevStatus.Mem.Total, DevStatus.Mem.Type, DevStatus.Mem.Hz)
	// CPU信息
	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_info"], prometheus.GaugeValue, 1, routerMac, strconv.Itoa(DevStatus.CPU.Core), DevStatus.CPU.Hz)

	// 启动时间
	uptime, _ := strconv.ParseFloat(DevStatus.UpTime, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["uptime"], prometheus.CounterValue, uptime, routerMac)
	// 历史设备数量
	ch <- prometheus.MustNewConstMetric(c.metrics["history_device_amount"], prometheus.GaugeValue, float64(DevStatus.Count.All), routerMac)
	// 在线设备数量
	ch <- prometheus.MustNewConstMetric(c.metrics["online_device_amount"], prometheus.GaugeValue, float64(DevStatus.Count.Online), routerMac)

	// CPU使用率
	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_load_percent"], prometheus.GaugeValue, DevStatus.CPU.Load*100, routerMac)
	// 内存使用率
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage_percent"], prometheus.CounterValue, DevStatus.Mem.Usage*100, routerMac)
	// 内存使用量
	totalMemory, _ := strconv.ParseFloat(strings.Split(DevStatus.Mem.Total, "MB")[0], 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage"], prometheus.GaugeValue, DevStatus.Mem.Usage*totalMemory, routerMac)
	// 温度
	ch <- prometheus.MustNewConstMetric(c.metrics["temperature"], prometheus.GaugeValue, float64(DevStatus.Temperature), routerMac)

	// WAN 当前上传速度
	uploadSpeed, _ := strconv.ParseFloat(DevStatus.Wan.Upspeed, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_speed"], prometheus.GaugeValue, uploadSpeed, routerMac)
	// WAN 当前下载速度
	downloadSpeed, _ := strconv.ParseFloat(DevStatus.Wan.Downspeed, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_speed"], prometheus.GaugeValue, downloadSpeed, routerMac)
	// WAN 上传流量
	uploadTraffic, _ := strconv.ParseFloat(DevStatus.Wan.Upload, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_traffic"], prometheus.CounterValue, uploadTraffic, routerMac)
	// WAN 下载流量
	downloadTraffic, _ := strconv.ParseFloat(DevStatus.Wan.Download, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_traffic"], prometheus.CounterValue, downloadTraffic, routerMac)
	// WAN 最大下载速度
	maxDownloadSpeed, _ := strconv.ParseFloat(DevStatus.Wan.Maxdownloadspeed, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["max_download_speed"], prometheus.GaugeValue, maxDownloadSpeed, routerMac)
	// WAN 最大上传速度
	maxUploadSpeed, _ := strconv.ParseFloat(DevStatus.Wan.Maxuploadspeed, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["max_upload_speed"], prometheus.GaugeValue, maxUploadSpeed, routerMac)

	for _, ipv4 := range WANInfo.Info.Ipv4 {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4"], prometheus.GaugeValue, 1, routerMac, ipv4.IP)
		mask, _ := SubNetMaskToLen(ipv4.Mask)
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4_mask"], prometheus.GaugeValue, float64(mask), routerMac, ipv4.IP)
	}
	for _, ipv6 := range WANInfo.Info.Ipv6Info.IP6Addr {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv6"], prometheus.GaugeValue, 1, routerMac, ipv6)
	}

	// WAN 信息
	details := WANInfo.Info.Details
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_info"], prometheus.GaugeValue, 1, routerMac, details.Username, details.Password, details.WanType)

	unknownAddressDeviceAmount := 0

	for _, dev := range DevStatus.Dev {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("存在不正常数据，请检查API", err)
			}
		}()

		var ip string
		for _, d := range Mactoip.List {
			if d.Mac == dev.Mac {
				ip = d.IP[0].IP
				break
			}
		}
		if ip == "" {
			ip = fmt.Sprintf("未知设备%d", unknownAddressDeviceAmount)
			unknownAddressDeviceAmount++
		}

		deviceMac := dev.Mac
		deviceName := dev.Devname

		// 设备上传流量
		deviceUploadTraffic := parseFloatValue(dev.Upload)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_traffic"], prometheus.CounterValue, deviceUploadTraffic, routerMac, deviceMac, deviceName, ip)
		// 设备下载流量
		deviceDownloadTraffic := parseFloatValue(dev.Download)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_download_traffic"], prometheus.CounterValue, deviceDownloadTraffic, routerMac, deviceMac, deviceName, ip)
		// 设备上传速度
		deviceUploadSpeed := parseFloatValue(dev.Upspeed)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_speed"], prometheus.GaugeValue, deviceUploadSpeed, routerMac, deviceMac, deviceName, ip)
		// 设备下载速度
		deviceDownloadSpeed := parseFloatValue(dev.Downspeed)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_download_speed"], prometheus.GaugeValue, deviceDownloadSpeed, routerMac, deviceMac, deviceName, ip)

		// 设备在线时间
		deviceOnlineTime := parseFloatValue(dev.Online)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_online_time"], prometheus.CounterValue, deviceOnlineTime, routerMac, deviceMac, deviceName, ip)

	}
}

func parseFloatValue(value interface{}) float64 {
	reflect.TypeOf(value).Name()
	switch value.(type) {
	case float64:
		return value.(float64)
	case string:
		value, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			fmt.Println("指标类型转换失败", err)
			value = 0
		}
		return value
	}
	return 0
}
