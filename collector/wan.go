package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/helloworlde/miwifi-exporter/config"
)

type WAN struct {
	Info struct {
		Mac     string `json:"mac"`
		Mtu     string `json:"mtu"`
		Details struct {
			Username string `json:"username"`
			Ifname   string `json:"ifname"`
			WanType  string `json:"wanType"`
			Service  string `json:"service"`
			Password string `json:"password"`
			Peerdns  string `json:"peerdns"`
		} `json:"details"`
		GateWay   string `json:"gateWay"`
		DNSAddrs1 string `json:"dnsAddrs1"`
		Status    int    `json:"status"`
		Uptime    int    `json:"uptime"`
		DNSAddrs  string `json:"dnsAddrs"`
		Ipv6Info  struct {
			WanType      string        `json:"wanType"`
			Ifname       string        `json:"ifname"`
			DNS          []interface{} `json:"dns"`
			IP6Addr      []string      `json:"ip6addr"`
			Peerdns      string        `json:"peerdns"`
			LanIP6Prefix []interface{} `json:"lan_ip6prefix"`
			LanIP6Addr   []interface{} `json:"lan_ip6addr"`
		} `json:"ipv6_info"`
		Ipv6Show int `json:"ipv6_show"`
		Link     int `json:"link"`
		Ipv4     []struct {
			Mask string `json:"mask"`
			IP   string `json:"ip"`
		} `json:"ipv4"`
	} `json:"info"`
	Code int `json:"code"`
}

var WANInfo WAN

func SubNetMaskToLen(netmask string) (int, error) {
	ipSplitArr := strings.Split(netmask, ".")
	if len(ipSplitArr) != 4 {
		return 0, fmt.Errorf("netmask:%v is not valid, pattern should like: 255.255.255.0", netmask)
	}
	ipv4MaskArr := make([]byte, 4)
	for i, value := range ipSplitArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("ipMaskToInt call strconv.Atoi error:[%v] string value is: [%s]", err, value)
		}
		if intValue > 255 {
			return 0, fmt.Errorf("netmask cannot greater than 255, current value is: [%s]", value)
		}
		ipv4MaskArr[i] = byte(intValue)
	}

	ones, _ := net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
	return ones, nil

}

func GetWAN() {
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/xqnetwork/wan_info",
		config.Config.IP, config.Token.Token))
	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(res.Body)
	log.Println("Get WAN: ", string(body))
	count := 0
	if err = json.Unmarshal(body, &WANInfo); err != nil {
		log.Println("Token失效，正在重试获取")
		config.GetConfig()
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误", err)
			os.Exit(1)
		}
	}
}
