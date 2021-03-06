package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/erikh/ldnsd/config"
	"github.com/erikh/ldnsd/proto"
	"github.com/erikh/ldnsd/service"
	"github.com/miekg/dns"
)

func init() {
	seed := os.Getenv("TEST_SEED")
	s, err := strconv.ParseInt(seed, 10, 64)
	if err != nil {
		s = time.Now().Unix()
		fmt.Println("Seed:", s)
	}
	rand.Seed(s)
	os.Remove("test.db")
}

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
	defer os.Remove("test.db")
	defer srv.Shutdown()

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

func randString(count, min int) string {
	s := []rune{}
	for i := 0; i < rand.Intn(count-min)+min; i++ {
		s = append(s, rune('a'+rand.Intn(26)))
	}

	return string(s)
}

func BenchmarkRecordInsert(b *testing.B) {
	srv, err := startService()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove("test.db")
	defer srv.Shutdown()

	client, err := proto.NewClient(config.DefaultGRPCListen, defaultCAFile, defaultCertFile, defaultKeyFile)
	if err != nil {
		b.Fatal(err)
	}

	hostChan := make(chan string, runtime.NumCPU()*2)
	hosts := map[string]struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		for {
		retry:
			select {
			case <-ctx.Done():
				return
			default:
			}

			host := randString(30, 3)
			if _, ok := hosts[host]; ok {
				goto retry
			}

			select {
			case <-ctx.Done():
				return
			case hostChan <- host:
				hosts[host] = struct{}{}
			}
		}
	}(ctx)

	b.Run("record insert", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if _, err := client.SetA(context.Background(), &proto.Record{Host: <-hostChan, Address: "1.2.3.4"}); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkRecordInsertThenQuery(b *testing.B) {
	srv, err := startService()
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove("test.db")
	defer srv.Shutdown()

	client, err := proto.NewClient(config.DefaultGRPCListen, defaultCAFile, defaultCertFile, defaultKeyFile)
	if err != nil {
		b.Fatal(err)
	}

	ip := net.ParseIP("1.2.3.4")

	// no buffer here otherwise we won't be able to resolve the on-buffer items that will get pushed to the map
	hostChan := make(chan string)
	hosts := map[string]struct{}{}

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
		retry:
			select {
			case <-ctx.Done():
				return
			default:
			}

			host := randString(30, 3)
			if _, ok := hosts[host]; ok {
				goto retry
			}

			select {
			case <-ctx.Done():
				return
			case hostChan <- host:
				hosts[host] = struct{}{}
			}
		}
	}(ctx)

	// BUG: not running this in parallel largely because I can't get the selects right above and it causes locking issues.
	b.Run("record insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := client.SetA(context.Background(), &proto.Record{Host: <-hostChan, Address: "1.2.3.4"}); err != nil {
				b.Fatal(err)
			}
		}
	})

	cancel()
	hostChan = make(chan string, runtime.NumCPU()*2)

	go func() {
		for {
			for key := range hosts {
				hostChan <- key
			}
		}
	}()

	b.Run("query", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m, err := msgClient(fmt.Sprintf("%s.internal.", <-hostChan))
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

func TestProtoRecordValidation(t *testing.T) {
	srv, err := startService()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test.db")
	defer srv.Shutdown()

	table := map[string]struct {
		r       *proto.Record
		success bool
	}{
		"basic": {
			r:       &proto.Record{Host: "test", Address: "127.0.0.1"},
			success: true,
		},
		"empty host": {
			r:       &proto.Record{Host: "", Address: "127.0.0.1"},
			success: false,
		},
		"empty ip": {
			r:       &proto.Record{Host: "test", Address: ""},
			success: false,
		},
		"bad ip": {
			r:       &proto.Record{Host: "test", Address: "abcdefgh"},
			success: false,
		},
		"ipv6 ip": {
			r:       &proto.Record{Host: "test", Address: "fe80::1"},
			success: false,
		},
		"invalid ipv4 ip": {
			r:       &proto.Record{Host: "test", Address: "256.1.1.1"},
			success: false,
		},
		"long string is looooooong": {
			r:       &proto.Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "fe80::1"},
			success: false,
		},
		"long string is too looooooong": {
			r:       &proto.Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "fe80::1"},
			success: false,
		},
		"long domain is looooooong": {
			r:       &proto.Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "127.0.0.1"},
			success: true,
		},
		"long domain has a really long part": {
			r:       &proto.Record{Host: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Address: "127.0.0.1"},
			success: false,
		},
	}

	client, err := proto.NewClient(config.DefaultGRPCListen, defaultCAFile, defaultCertFile, defaultKeyFile)
	if err != nil {
		t.Fatal(err)
	}

	for testName, result := range table {
		_, resultErr := client.SetA(context.Background(), result.r)
		if result.success && resultErr != nil {
			t.Fatalf("Result for %q should be success but was %v", testName, resultErr)
		}
		if !result.success && resultErr == nil {
			t.Fatalf("Result for %q should NOT be success but was.", testName)
		}
	}
}
