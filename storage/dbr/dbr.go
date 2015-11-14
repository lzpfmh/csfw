package dbr

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gocraft/dbr/dialect"
)

// Open instantiates a Connection for a given database/sql connection
// and event receiver
func Open(dsn string, opts ...ConnectionOption) (c *Connection, err error) {
	c = &Connection{
		Dialect:       dialect.MySQL,
		EventReceiver: nullReceiver,
	}
	c.ApplyOpts(opts...)

	if c.DB != nil {
		return c, nil
	}

	c.DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ApplyOpts applies options to a connection
func (c *Connection) ApplyOpts(opts ...ConnectionOption) *Connection {
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// Connection is a connection to the database with an EventReceiver
// to send events, errors, and timings to
type Connection struct {
	*sql.DB
	Dialect Dialect
	EventReceiver
}

// Session represents a business unit of execution for some connection
type Session struct {
	*Connection
	EventReceiver
}

// ConnectionOption can be used as an argument in NewConnection to configure a connection.
type ConnectionOption func(c *Connection)

// SetDB sets the DB value to a connection. If set ignores the DSN values.
func SetDB(db *sql.DB) ConnectionOption {
	if db == nil {
		panic("DB argument cannot be nil")
	}
	return func(c *Connection) {
		c.DB = db
	}
}

// SetEventReceiver sets the event receiver for a connection.
func SetEventReceiver(log EventReceiver) ConnectionOption {
	if log == nil {
		log = nullReceiver
	}
	return func(c *Connection) {
		c.EventReceiver = log
	}
}

// SessionOption can be used as an argument in NewSession to configure a session.
type SessionOption func(cxn *Connection, s *Session) SessionOption

// SetSessionEventReceiver sets an event receiver securely to a session. Falls
// back to the parent event receiver if argument is nil.
// This function adheres http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func SetSessionEventReceiver(log EventReceiver) SessionOption {
	return func(cxn *Connection, s *Session) SessionOption {
		previous := s.EventReceiver
		if log == nil {
			log = cxn.EventReceiver // Use parent instrumentation
		}
		s.EventReceiver = log
		return SetSessionEventReceiver(previous)
	}
}

// NewSession instantiates a Session for the Connection
func (c *Connection) NewSession(opts ...SessionOption) *Session {
	s := &Session{
		Connection:    c,
		EventReceiver: c.EventReceiver, // Use parent instrumentation
	}
	s.ApplyOpts(opts...)
	return s
}

// Close closes the database, releasing any open resources.
func (c *Connection) Close() error {
	return c.EventErr("dbr.connection.close", c.DB.Close())
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
func (c *Connection) Ping() error {
	return c.EventErr("dbr.connection.ping", c.DB.Ping())
}

// NewSession instantiates a Session for the Connection
func (s *Session) ApplyOpts(opts ...SessionOption) (previous SessionOption) {
	for _, opt := range opts {
		if opt != nil {
			previous = opt(s.Connection, s)
		}
	}
	return previous
}

// MustOpenAndVerify is like NewConnection but it verifies the connection
// and panics on errors.
func MustOpenAndVerify(dsn string, opts ...ConnectionOption) *Connection {
	c, err := Open(dsn, opts...)
	if err != nil {
		panic(err)
	}
	if err := c.Ping(); err != nil {
		panic(err)
	}
	return c
}

// Ensure that tx and session are session runner
var (
	_ SessionRunner = (*Tx)(nil)
	_ SessionRunner = (*Session)(nil)
)

// SessionRunner can do anything that a Session can except start a transaction.
type SessionRunner interface {
	Select(column ...string) *SelectBuilder
	SelectBySql(query string, value ...interface{}) *SelectBuilder

	InsertInto(table string) *InsertBuilder
	InsertBySql(query string, value ...interface{}) *InsertBuilder

	Update(table string) *UpdateBuilder
	UpdateBySql(query string, value ...interface{}) *UpdateBuilder

	DeleteFrom(table string) *DeleteBuilder
	DeleteBySql(query string, value ...interface{}) *DeleteBuilder
}

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type builder interface {
	ToSql() (string, []interface{})
}

func exec(runner runner, log EventReceiver, builder builder, d Dialect) (sql.Result, error) {
	query, value := builder.ToSql()
	query, err := InterpolateForDialect(query, value, d)
	if err != nil {
		return nil, log.EventErrKv("dbr.exec.interpolate", err, kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		})
	}

	startTime := time.Now()
	defer func() {
		log.TimingKv("dbr.exec", time.Since(startTime).Nanoseconds(), kvs{
			"sql": query,
		})
	}()

	result, err := runner.Exec(query)
	if err != nil {
		return result, log.EventErrKv("dbr.exec.exec", err, kvs{
			"sql": query,
		})
	}
	return result, nil
}

func query(runner runner, log EventReceiver, builder builder, d Dialect, v interface{}) (int, error) {
	query, value := builder.ToSql()
	query, err := InterpolateForDialect(query, value, d)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.interpolate", err, kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		})
	}

	startTime := time.Now()
	defer func() {
		log.TimingKv("dbr.select", time.Since(startTime).Nanoseconds(), kvs{
			"sql": query,
		})
	}()

	rows, err := runner.Query(query)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.load.query", err, kvs{
			"sql": query,
		})
	}
	count, err := Load(rows, v)
	if err != nil {
		return 0, log.EventErrKv("dbr.select.load.scan", err, kvs{
			"sql": query,
		})
	}
	return count, nil
}
