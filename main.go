package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	logFmt = "[%s] %s\n"
)

var (
	resv *net.Resolver

	theClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    100,
			IdleConnTimeout: 5 * time.Minute,
			MaxConnsPerHost: 30,
		},
		Timeout: 10 * time.Second,
	}
)

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}

func log(line string, args ...interface{}) {
	s := line
	if len(args) > 0 {
		s = fmt.Sprintf(line, args)
	}
	fmt.Printf(logFmt, timeStr(), s)
}

func timeStr() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

func main() {
	var srvAddr string
	var srvPort int
	var useTLS bool
	flag.StringVar(&srvAddr, "hostname", "", "Hostname of the Flix server")
	flag.IntVar(&srvPort, "port", 0, "Port of the Flix server")
	flag.BoolVar(&useTLS, "use-tls", false, "Use TLS when connecting to the Flix server")
	flag.Parse()

	if srvAddr == "" {
		log("Hostname is required. This should be the hostname of the Flix server, not including 'http(s)://'")
		os.Exit(1)
	}
	if srvPort == 0 {
		log("Port is required. This should be the port of the Flix server, not including ':'")
		os.Exit(1)
	}
	log("Address to lookup: %s", srvAddr)
	log("Port to use: %d", srvPort)
	if useTLS {
		log("Using TLS")
	} else {
		log("Not using TLS")
	}

	log("Looking up address names")
	c, cancel := ctx()
	names, err := resv.LookupAddr(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up names: %v", err)
	} else {
		log("Got names: %s", strings.Join(names, ", "))
	}

	log("Looking up CNAME")
	c, cancel = ctx()
	cname, err := resv.LookupCNAME(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up names: %v", err)
	} else {
		log("Got CNAME: %v", cname)
	}

	log("Looking up host addresses")
	c, cancel = ctx()
	addrs, err := resv.LookupHost(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up host addresses: %v", err)
	} else {
		log("Got addresses: %v", strings.Join(addrs, ", "))
	}

	log("Looking up IP addresses")
	c, cancel = ctx()
	ips, err := resv.LookupIPAddr(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up IP addresses: %v", err)
	} else {
		ipStrs := make([]string, len(ips))
		for i := range ips {
			ipStrs[i] = ips[i].String()
		}
		log("Got IP addresses: %s", strings.Join(ipStrs, ", "))
	}

	log("Looking up MX records")
	c, cancel = ctx()
	mxs, err := resv.LookupMX(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up MX records: %v", err)
	} else {
		mxStrs := make([]string, len(mxs))
		for i := range mxs {
			mxStrs[i] = fmt.Sprintf("%s:%d", mxs[i].Host, mxs[i].Pref)
		}
		log("Got MX records: %s", strings.Join(mxStrs, ", "))
	}

	log("Looking up NS records")
	c, cancel = ctx()
	nss, err := resv.LookupNS(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up NS records: %v", err)
	} else {
		nsStrs := make([]string, len(nss))
		for i := range nss {
			nsStrs[i] = nss[i].Host
		}
		log("Got NS records: %s", strings.Join(nsStrs, ", "))
	}

	log("Looking up TXT records")
	c, cancel = ctx()
	txts, err := resv.LookupTXT(c, srvAddr)
	cancel()
	if err != nil {
		log("Could not look up TXTs: %v", err)
	} else {
		log("Got TXT records: %s", strings.Join(txts, ", "))
	}

	log("Attempting to connect to the Flix server public Info page")
	proto := "http://"
	if useTLS {
		proto = "https://"
	}
	url := fmt.Sprintf("%s%s:%d/info", proto, srvAddr, srvPort)
	log("URL: %s", url)
	resp, err := theClient.Get(url)
	if err != nil {
		log("Failed to get info page: %v", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log("Failed to read response: %v", err)
		os.Exit(1)
	}
	_ = resp.Body.Close()
	log("Response body:")
	log(string(body[:]))
}
