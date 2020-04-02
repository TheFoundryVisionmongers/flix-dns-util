package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bcvery1/gotime"
)

const (
	logFmt = "[%s] %s\n"
)

var (
	resv    *net.Resolver
	srvAddr string
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
	return gotime.FormatDateTime(time.Now(), "%Y/%m/%d %H:%M:%S")
}

func main() {
	flag.StringVar(&srvAddr, "hostname", "", "Hostname of the Flix server")
	flag.Parse()

	if srvAddr == "" {
		log("Hostname is required. This should be the hostname of the Flix server, not including 'http(s)://' or the port ':8080'")
		os.Exit(1)
	}
	log("Address to lookup: %s", srvAddr)

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
}
