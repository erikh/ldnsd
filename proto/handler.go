package proto

import (
	"code.hollensbe.org/erikh/ldnsd/dnsdb"
	"github.com/erikh/dnsserver/db"
	grpc "google.golang.org/grpc"
)

// Handler is the control plane handler.
type Handler struct {
	db db.DB
}

// Boot boots the grpc service
func Boot(db db.DB) *grpc.Server {
	h := &Handler{db: db}

	s := grpc.NewServer()
	RegisterDNSControlServer(s, h)

	return s
}

func toGRPC(record *dnsdb.Record) *Record {
	return &Record{
		Host:    record.Host,
		Address: record.Address,
	}
}
