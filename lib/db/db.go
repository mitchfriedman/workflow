package database

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
)

type DB struct {
	Master *gorm.DB
	Reader *gorm.DB
}

func (d *DB) Close() error {
	if err := d.Master.Close(); err != nil {
		return errors.Wrap(err, "failed to close master")
	}

	if err := d.Reader.Close(); err != nil {
		return errors.Wrap(err, "failed to close reader")
	}

	return nil
}

func configureDBArgs(url string, timeoutMS int) string {
	if strings.HasSuffix(url, "sslmode=disable") {
		url = url + "&"
	} else if !strings.HasSuffix(url, "?") {
		url = url + "?"
	}

	url = url + fmt.Sprintf("statement_timeout=%d", timeoutMS)

	return url
}

func Connect(masterURL, readerURL string, logQueries bool) (*DB, error) {
	mURL := configureDBArgs(masterURL, 30000)
	writer, err := gorm.Open("postgres", mURL)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to master Postgres")
	}

	rURL := configureDBArgs(readerURL, 30000)
	reader, err := gorm.Open("postgres", rURL)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to reader Postgres")
	}

	writer.LogMode(logQueries)
	reader.LogMode(logQueries)

	return &DB{
		Master: writer,
		Reader: reader,
	}, nil
}
