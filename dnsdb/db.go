package dnsdb

import (
	"net"
	"regexp"
	"strings"

	dnsserverDB "github.com/erikh/dnsserver/db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // import sqlite3
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var hostMatch = regexp.MustCompile(`^[a-z][0-9a-z-]{0,62}$`)

var (
	// ErrNotSupported is for when something is not supported by this interface
	ErrNotSupported = errors.New("not supported")
)

// DB is the outer shell for the gorm DB handle.
type DB struct {
	db *gorm.DB
}

// New opens the DB
func New(dbfile string) (dnsserverDB.DB, error) {
	db, err := gorm.Open("sqlite3", dbfile)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to db")
	}

	if err := db.AutoMigrate(&Record{}).Error; err != nil {
		return nil, errors.Wrap(err, "while migrating database")
	}

	return &DB{db: db}, nil
}

// Close the database
func (db *DB) Close() error {
	return db.db.Close()
}

// Record is the notion of an A record in the database.
type Record struct {
	Host    string `gorm:"primary_key"`
	Address string
}

// Validate ensures the record is safe to insert.
func (r *Record) Validate() error {
	ip := net.ParseIP(r.Address)
	if len(ip) == 0 {
		return errors.New("IP address did not parse")
	}

	if !ip.To4().Equal(ip) {
		return errors.New("IP is not IPv4. ldnsd does not support IPv6 yet")
	}

	return r.validateHost()
}

func (r *Record) validateHost() error {
	if len(r.Host) == 0 {
		return errors.New("name is 0 length")
	}

	if len(r.Host) > 255 {
		return errors.New("name is longer than 255 characters")
	}

	for _, name := range strings.Split(r.Host, ".") {
		if !hostMatch.MatchString(name) {
			return errors.New("names in DNS must be 63 characters or less, per part")
		}
	}

	return nil
}

// IP returns the parsed IP address of the record in IPv4 32-bit format.
func (r *Record) IP() net.IP {
	return net.ParseIP(r.Address).To4()
}

// SetA sets an A record in the database.
func (db *DB) SetA(host string, ip net.IP) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		r := &Record{
			Host:    host,
			Address: ip.String(),
		}

		if err := r.Validate(); err != nil {
			return errors.Wrap(err, "during record validation")
		}

		return tx.Create(r).Error
	})
}

// GetA retrieves an A record in the database.
func (db *DB) GetA(host string) (net.IP, error) {
	r := &Record{}

	err := db.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(r, "host = ?", host).Error
	})

	if err := r.Validate(); err != nil {
		return nil, errors.Wrap(err, "during validation of record fetched")
	}

	return r.IP(), err
}

// DeleteA removes a DNS record
func (db *DB) DeleteA(host string) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		r := &Record{Host: host}
		if err := r.validateHost(); err != nil {
			return errors.Wrap(err, "during validation of hostname")
		}

		return tx.Delete(r).Error
	})
}

// ListA lists all the A records in the table
func (db *DB) ListA() (dnsserverDB.ARecords, error) {
	tmp := dnsserverDB.ARecords{}

	return tmp, db.db.Transaction(func(tx *gorm.DB) error {
		recs := []*Record{}
		if err := tx.Find(&recs).Error; err != nil {
			return err
		}

		for _, rec := range recs {
			if err := rec.Validate(); err != nil {
				logrus.Errorf("Error validating record %q/%q during database traversal in list function: %v. Skipping record; please file an issue.", rec.Host, rec.IP(), err)
				continue
			}

			tmp[rec.Host] = rec.IP()
		}

		return nil
	})
}

// ListSRV does nothing but fulfill an interface.
func (db *DB) ListSRV() (dnsserverDB.SRVRecords, error) {
	return nil, ErrNotSupported
}

// SetSRV does nothing but fulfill an interface.
func (db *DB) SetSRV(string, *dnsserverDB.SRVRecord) error {
	return ErrNotSupported
}

// GetSRV does nothing but fulfill an interface.
func (db *DB) GetSRV(string) (*dnsserverDB.SRVRecord, error) {
	return nil, ErrNotSupported
}

// DeleteSRV does nothing but fulfill an interface.
func (db *DB) DeleteSRV(string) error { return ErrNotSupported }
