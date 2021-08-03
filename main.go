package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	pollingDuration time.Duration = 20 * time.Second
	listeningPort                 = ":9898"
)

var (
	mu            sync.RWMutex
	ethereumNodes []string
)

func resolveDNSAddress(addressRecord []string) ([]string, error) {
	// Resolve DNS A record to a set of IP addresses
	ipAddresses, err := ResolveAddressRecord(addressRecord)
	if err != nil {
		log.Errorf("Error resolving DNS address record: %s", err)
		return nil, err
	}

	log.Printf("%s resolved to %s", addressRecord, ipAddresses)
	return ipAddresses, nil
}

func updateEthereumNodes(ipAddresses []string) {
	log.Printf("updateEthereumNodes ip addr received: %+v", ipAddresses)

	// Retrieve enode from each IP address
	for _, ipAddress := range ipAddresses {
		ip := ipAddress
		ipAddrPort := "8080"

		ipAddrSplit := strings.Split(ipAddress, ":")
		if len(ipAddrSplit) == 2 {
			ip = ipAddrSplit[0]
			ipAddrPort = ipAddrSplit[1]
		}

		resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, ipAddrPort))
		if err != nil {
			log.Errorf("Error retrieving enode address: %s", err)
			continue
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Error parsing response: %s", err)
			continue
		}

		var enodeAddress = strings.TrimSpace(string(contents))
		writeAddrEthNode(enodeAddress)
		log.Infof("%s with enode address %s", ipAddress, enodeAddress)
	}

}

func writeAddrEthNode(enodeAddress string) {
	mu.Lock()
	defer mu.Unlock()

	for _, enodeAddr := range ethereumNodes {
		if enodeAddr == enodeAddress {
			return
		}
	}

	ethereumNodes = append(ethereumNodes, enodeAddress)
}

func startPollUpdateEthereumNodes(addressRecord []string) {
	for {
		ipAddress, err := resolveDNSAddress(addressRecord)
		if err != nil {
			return
		}

		go updateEthereumNodes(ipAddress)
		<-time.After(pollingDuration)
	}
}

func startPollUpdateEthereumNodesIp(bootnodeIPs []string) {
	for {
		go updateEthereumNodes(bootnodeIPs)
		<-time.After(pollingDuration)
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("handling request from %s", r.RemoteAddr)

	mu.RLock()
	defer mu.RUnlock()

	fmt.Fprintln(w, strings.Join(ethereumNodes, ","))
}

func main() {
	bootNodeService := flag.String("service", os.Getenv("BOOTNODE_SERVICE"), "Comma separated for multiple DNS A Record for `bootnode` services. Alternatively set `BOOTNODE_SERVICE` env.")
	bootNodeIps := flag.String("ips", os.Getenv("BOOTNODE_IPS"), "Comma separated for multiple `bootnode` IP. Alternatively set `BOOTNODE_IPS` env.")
	flag.Parse()

	if *bootNodeService == "" && *bootNodeIps == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	bootNodeServiceStr := *bootNodeService
	arrBootNodeServices := strings.Split(bootNodeServiceStr, ",")

	if len(arrBootNodeServices) > 0 && *bootNodeService != "" {
		log.Infof("starting bootnode-registrar DNS: %s.", *bootNodeService)
		go startPollUpdateEthereumNodes(arrBootNodeServices)
	}

	arrBootNodeIPs := strings.Split(*bootNodeIps, ",")
	if len(arrBootNodeIPs) > 0 && *bootNodeIps != "" {
		log.Infof("starting bootnode-registrar IP: %s.", *bootNodeIps)
		go startPollUpdateEthereumNodesIp(arrBootNodeIPs)
	}

	log.Infof("Start web handler at: %s.", listeningPort)
	http.HandleFunc("/", webHandler)
	log.Fatal(http.ListenAndServe(listeningPort, nil))
}
