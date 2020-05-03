package dnsdb

import (
	"net"

	"github.com/erikh/dnsserver/db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // import sqlite3
	"github.com/pkg/errors"
)

var (
	// ErrNotSupported is for when something is not supported by this interface
	ErrNotSupported = errors.New("not supported")
)

// DB is the outer shell for the gorm DB handle.
type DB struct {
	db *gorm.DB
}

// New opens the DB
func New(dbfile string) (db.DB, error) {
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

// IP returns the parsed IP address of the record in IPv4 32-bit format.
func (r *Record) IP() net.IP {
	return net.ParseIP(r.Address).To4()
}

// SetA sets an A record in the database.
func (db *DB) SetA(host string, ip net.IP) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&Record{
			Host:    host,
			Address: ip.String(),
		}).Error
	})
}

// GetA retrieves an A record in the database.
func (db *DB) GetA(host string) (net.IP, error) {
	r := &Record{}

	err := db.db.Transaction(func(tx *gorm.DB) error {
		return tx.First(r, "host = ?", host).Error
	})

	return r.IP(), err
}

// DeleteA removes a DNS record
func (db *DB) DeleteA(host string) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&Record{Host: host}).Error
	})
}

// ListA lists all the A records in the table
func (db *DB) ListA() (map[string]net.IP, error) {
	tmp := map[string]net.IP{}

	return tmp, db.db.Transaction(func(tx *gorm.DB) error {
		recs := []*Record{}
		if err := tx.Find(&recs).Error; err != nil {
			return err
		}

		for _, rec := range recs {
			tmp[rec.Host] = rec.IP()
		}

		return nil
	})
}

// ListSRV does nothing but fulfill an interface.
func (db *DB) ListSRV() (map[string]*db.SRVRecord, error) {
	return nil, ErrNotSupported
}

// SetSRV does nothing but fulfill an interface.
func (db *DB) SetSRV(string, *db.SRVRecord) error {
	return ErrNotSupported
}

// GetSRV does nothing but fulfill an interface.
func (db *DB) GetSRV(string) (*db.SRVRecord, error) {
	return nil, ErrNotSupported
}

// DeleteSRV does nothing but fulfill an interface.
func (db *DB) DeleteSRV(string) error { return ErrNotSupported }
