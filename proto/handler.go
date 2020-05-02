package proto

import (
	context "context"

	"code.hollensbe.org/erikh/ldnsd/dnsdb"
	"github.com/erikh/dnsserver/db"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

func fromGRPC(record *Record) *dnsdb.Record {
	return &dnsdb.Record{
		Host:    record.Host,
		Address: record.Address,
	}
}

func toGRPC(record *dnsdb.Record) *Record {
	return &Record{
		Host:    record.Host,
		Address: record.Address,
	}
}

// SetA sets a new A record.
func (h *Handler) SetA(ctx context.Context, record *Record) (*empty.Empty, error) {
	r := fromGRPC(record)

	if err := h.db.SetA(r.Host, r.IP()); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Aborted, "%v", err)
	}

	return &empty.Empty{}, nil
}

// DeleteA removes an existing A record
func (h *Handler) DeleteA(ctx context.Context, record *Record) (*empty.Empty, error) {
	r := fromGRPC(record)

	h.db.DeleteA(r.Host)

	return &empty.Empty{}, nil
}
