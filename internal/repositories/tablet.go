package repositories

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Tablet struct {
	ID       int64     `db:"id"`
	IP       string    `db:"ip"`
	Name     string    `db:"name"`
	Version  string    `db:"version"`
	Online   bool      `db:"online"`
	LastSeen time.Time `db:"last_seen"`
}

type TabletRepository interface {
	InitTable() error
	Save(t *Tablet) error
	GetAll() ([]Tablet, error)
	GetByID(id int64) (*Tablet, error)
}

type sqliteTabletRepo struct {
	db *sqlx.DB
}

func NewTabletRepository(db *sqlx.DB) TabletRepository {
	return &sqliteTabletRepo{db: db}
}

func (r *sqliteTabletRepo) InitTable() error {
	query := `CREATE TABLE IF NOT EXISTS tablets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip TEXT NOT NULL UNIQUE,
		name TEXT,
		version TEXT,
		online BOOLEAN DEFAULT 0,
		last_seen DATETIME
	);`
	_, err := r.db.Exec(query)
	return err
}

func (r *sqliteTabletRepo) Save(t *Tablet) error {
	if t.LastSeen.IsZero() {
		t.LastSeen = time.Now()
	}

	query := `INSERT INTO tablets (id, ip, name, version, online, last_seen)
        VALUES (NULLIF(:id, 0), :ip, :name, :version, :online, :last_seen)
        ON CONFLICT(id) DO UPDATE SET
            ip=excluded.ip,
            name=excluded.name,
            version=excluded.version,
            online=excluded.online,
            last_seen=excluded.last_seen
        ON CONFLICT(ip) DO UPDATE SET
            name=excluded.name,
            version=excluded.version,
            online=excluded.online,
            last_seen=excluded.last_seen`

	_, err := r.db.NamedExec(query, t)
	return err
}

func (r *sqliteTabletRepo) GetAll() ([]Tablet, error) {
	var tablets []Tablet
	err := r.db.Select(&tablets, "SELECT * FROM tablets")
	return tablets, err
}

func (r *sqliteTabletRepo) GetByID(id int64) (*Tablet, error) {
	var t Tablet
	err := r.db.Get(&t, "SELECT * FROM tablets WHERE id = ?", id)
	return &t, err
}
