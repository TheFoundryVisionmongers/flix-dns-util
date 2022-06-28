package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/TheFoundryVisionmongers/flix-dns-util/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
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
		s = fmt.Sprintf(line, args...)
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
	var grpcPort int
	flag.StringVar(&srvAddr, "hostname", "", "Hostname of the Flix server")
	flag.IntVar(&srvPort, "port", 0, "Port of the Flix server")
	flag.IntVar(&grpcPort, "transfer-port", 0, "File transfer port of the Flix server")
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
	if grpcPort == 0 {
		log("Transfer port is required. This should be the port used to transfer files to the Flix server, not including ':'")
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
	infoUrl := fmt.Sprintf("%s%s:%d/info", proto, srvAddr, srvPort)
	log("URL: %s", infoUrl)
	resp, err := theClient.Get(infoUrl)
	if err != nil {
		log("Failed to get info page: %v", err)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log("Failed to read response: %v", err)
		} else {
			_ = resp.Body.Close()
			log("Response body:")
			log(string(body[:]))
		}
	}

	grpcAddress := net.JoinHostPort(srvAddr, strconv.Itoa(grpcPort))

	log("Attempting to get proxy information")
	proxyUrl, err := http.ProxyFromEnvironment(&http.Request{
		URL: &url.URL{
			Scheme: "https",
			Host:   grpcAddress,
		},
	})
	if err != nil {
		log("Could not get proxy information: %v", err)
	} else if proxyUrl != nil {
		log("Got proxy URL: %s", proxyUrl.String())
	} else {
		log("No proxy discovered")
	}

	testGrpc("with default options", grpcAddress)
	testGrpc("with no proxy", grpcAddress, grpc.WithNoProxy())
}

func testGrpc(name string, grpcAddress string, options ...grpc.DialOption) {
	log("Connecting to %s over gRPC %s", grpcAddress, name)

	grpcCreds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
	})
	interceptor := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log("Calling gRPC method %s on %s", method, cc.Target())
		log("State: %v", cc.GetState())
		return streamer(ctx, desc, cc, method, opts...)
	}

	c, cancel := ctx()
	options = append(options, grpc.WithReturnConnectionError(), grpc.WithTransportCredentials(grpcCreds), grpc.WithStreamInterceptor(interceptor))
	conn, err := grpc.DialContext(c, grpcAddress, options...)
	cancel()
	if err != nil {
		log("Failed to dial server: %v", err)
	} else {
		client := pb.NewFileTransferClient(conn)

		var header metadata.MD
		var trailer metadata.MD
		var peerInfo peer.Peer
		log("Initiating transfer request")
		c, cancel = ctx()
		transferStream, err := client.Transfer(c, grpc.Header(&header), grpc.Trailer(&trailer), grpc.Peer(&peerInfo))
		if err != nil {
			log("Failed to call Transfer(): %v", err)
		} else {
			log("Sending transfer request")
			if err = transferStream.Send(&emptypb.Empty{}); err != nil {
				log("Failed to send transfer request: %v", err)
			} else {
				log("Receiving transfer response")
				_, err = transferStream.Recv()
				if err != nil && !strings.Contains(err.Error(), "FNAUTH signature not set") {
					log("Failed to receive transfer response: %v", err)
				}
				log("Got transfer response")
			}
		}
		cancel()

		log("gRPC header: %v", header)
		log("gRPC trailer: %v", trailer)
		if peerInfo.Addr != nil {
			log("gRPC peer information: %s %s", peerInfo.Addr.String(), peerInfo.Addr.Network())
		}
	}
}
