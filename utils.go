package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const defaultTimeout = 333 * time.Millisecond
const regxIP = `(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)`

var ipAPI = []string{
	"http://ip.cn", "http://ipinfo.io", "http://ifconfig.co", "http://myip.ipip.net",
	"http://cnc.synology.cn:81", "http://jpc.synology.com:81", "http://usc.synology.com:81",
	"http://ip.6655.com/ip.aspx", "http://pv.sohu.com/cityjson?ie=utf-8", "http://whois.pconline.com.cn/ipJson.jsp",
}

func getIP() (ip string) {
	var (
		wg    sync.WaitGroup
		lc    sync.Mutex
		ipMap = make(map[string]int, len(ipAPI))
	)
	for _, url := range ipAPI {
		wg.Add(1)
		go func(url string) {
			ip := regexp.MustCompile(regxIP).FindString(wGet(url, defaultTimeout))
			// log.Println(ip, url)
			if len(ip) > 0 {
				lc.Lock()
				ipMap[ip]++
				lc.Unlock()
			}
			wg.Done()
		}(url)
	}
	wg.Wait()
	max := 0
	for k, v := range ipMap {
		if v > len(ipAPI)/2 {
			return k
		} else if v > max {
			max = v
			ip = k
		}
	}

	if len(ip) == 0 {
		// Use First ipAPI as failsafe
		ip = regexp.MustCompile(regxIP).FindString(wGet(ipAPI[0], 20*defaultTimeout))
	}
	return
}

func wGet(url string, timeout time.Duration) (str string) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	str = string(body)
	return
}
