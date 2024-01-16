package repository

import (
	"database/sql"
	"time"
)

type Dependencies struct {
	DB *sql.DB

	NowFunc func() time.Time
}
