package main

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"code.hollensbe.org/erikh/ldnsd/config"
	"code.hollensbe.org/erikh/ldnsd/proto"
	"code.hollensbe.org/erikh/ldnsd/service"
	"github.com/miekg/dns"
)

const (
	defaultCAFile    = "/etc/ldnsd/rootCA.pem"
	defaultCertFile  = "/etc/ldnsd/client.pem"
	defaultKeyFile   = "/etc/ldnsd/client.key"
	defaultDNSListen = "127.0.0.1:5300"
)

func msgClient(fqdn string) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeA)
	return dns.Exchange(m, defaultDNSListen)
}

func startService() (*service.Service, error) {
	c := config.Empty()
	c.DBFile = "test.db"
	c.DNSListen = defaultDNSListen

	srv, err := service.New("test-ldnsd", c)
	if err != nil {
		return nil, err
	}

	go srv.Boot()
	time.Sleep(100 * time.Millisecond)

	return srv, nil
}

func BenchmarkDNSSingleDomain(b *testing.B) {
	srv, err := startService()
	if err != nil {
		b.Fatal(err)
	}
	defer srv.Shutdown()
	defer os.Remove("test.db")

	client, err := proto.NewClient(config.DefaultGRPCListen, defaultCAFile, defaultCertFile, defaultKeyFile)
	if err != nil {
		b.Fatal(err)
	}

	if _, err := client.SetA(context.Background(), &proto.Record{Host: "test", Address: "1.2.3.4"}); err != nil {
		b.Fatal(err)
	}

	ip := net.ParseIP("1.2.3.4")

	b.Run("test queries", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m, err := msgClient("test.internal.")
				if err != nil {
					b.Log(err)
					continue
				}

				aRecord := m.Answer[0].(*dns.A).A
				if !aRecord.Equal(ip) {
					b.Fatalf("IP %q does not match registered IP %q", aRecord, ip)
				}
			}
		})
	})
}
