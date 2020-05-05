package proto

import (
	context "context"

	"github.com/erikh/dnsserver"
	"github.com/erikh/ldnsd/dnsdb"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Handler is the control plane handler.
type Handler struct {
	srv *dnsserver.Server
}

// Boot boots the grpc service
func Boot(srv *dnsserver.Server) *grpc.Server {
	h := &Handler{srv: srv}

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

// SetA sets a new A record.
func (h *Handler) SetA(ctx context.Context, record *Record) (*empty.Empty, error) {
	r := fromGRPC(record)

	if err := h.srv.SetA(r.Host, r.IP()); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Aborted, "%v", err)
	}

	return &empty.Empty{}, nil
}

// DeleteA removes an existing A record
func (h *Handler) DeleteA(ctx context.Context, record *Record) (*empty.Empty, error) {
	r := fromGRPC(record)

	if err := h.srv.DeleteA(r.Host); err != nil {
		return &empty.Empty{}, status.Errorf(codes.Aborted, "%v", err)
	}

	return &empty.Empty{}, nil
}

// ListA returns a list of DNS records that the database is currently holding.
func (h *Handler) ListA(ctx context.Context, empty *empty.Empty) (*Records, error) {
	m, err := h.srv.ListA()
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "%v", err)
	}

	records := &Records{}
	for name, ip := range m {
		records.Records = append(records.Records, &Record{Host: name, Address: ip.String()})
	}

	return records, nil
}
