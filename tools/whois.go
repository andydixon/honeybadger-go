package tools

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const (
	// IANA_WHOIS_SERVER is iana whois server
	IANA_WHOIS_SERVER = "whois.iana.org"
	// DOMAIN_WHOIS_SERVER is tld whois server
	DOMAIN_WHOIS_SERVER = "whois-servers.net"
	// WHOIS_PORT is default whois port
	WHOIS_PORT = "43"
)

// Whois do the whois query and returns whois info
func Whois(domain string) (result string, err error) {
	domain = strings.Trim(strings.TrimSpace(domain), ".")
	if domain == "" {
		err = fmt.Errorf("Domain is empty")
		return
	}

	var servers []string
	result, err = query(domain, servers...)
	if err != nil {
		return
	}

	token := "Registrar WHOIS Server:"
	if IsIpv4(domain) {
		token = "whois:"
	}

	start := strings.Index(result, token)
	if start == -1 {
		return
	}

	start += len(token)
	end := strings.Index(result[start:], "\n")
	server := strings.TrimSpace(result[start : start+end])
	if server == "" {
		return
	}

	tmpResult, err := query(domain, server)
	if err != nil {
		return
	}

	result += tmpResult

	return
}

// query do the query
func query(domain string, servers ...string) (result string, err error) {
	server := IANA_WHOIS_SERVER
	if len(servers) == 0 || servers[0] == "" {
		if !IsIpv4(domain) {
			domains := strings.Split(domain, ".")
			if len(domains) > 1 {
				ext := domains[len(domains)-1]
				if strings.Contains(ext, "/") {
					ext = strings.Split(ext, "/")[0]
				}
				server = ext + "." + DOMAIN_WHOIS_SERVER
			}
		}
	} else {
		server = strings.ToLower(servers[0])
		if server == "whois.arin.net" {
			domain = "n + " + domain
		}
	}

	conn, e := net.DialTimeout("tcp", net.JoinHostPort(server, WHOIS_PORT), time.Second*30)
	if e != nil {
		err = e
		return
	}

	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	_, err = conn.Write([]byte(domain + "\r\n"))
	if err != nil {
		return
	}

	buffer, err := ioutil.ReadAll(conn)
	if err != nil {
		return
	}

	result = string(buffer)

	return
}

// IsIpv4 returns string is an ipv4 ip
func IsIpv4(ip string) bool {
	i := net.ParseIP(ip)
	return i.To4() != nil
}
