package repositories

import (
	"github.com/jmoiron/sqlx"
)

type Group struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Color       string `db:"color"`
}

type GroupRepository interface {
	InitTable() error
	// Group CRUD
	Create(g *Group) (int64, error)
	GetAll() ([]Group, error)
	GetByID(id int64) (*Group, error)
	Update(g *Group) error
	Delete(id int64) error

	// Relations
	AddTabletToGroup(tabletID, groupID int64) error
	RemoveTabletFromGroup(tabletID, groupID int64) error
	GetGroupsByTablet(tabletID int64) ([]Group, error)
	GetTabletsByGroup(groupID int64) ([]Tablet, error) // Ajouté
}

type sqliteGroupRepo struct {
	db *sqlx.DB
}

func NewGroupRepository(db *sqlx.DB) GroupRepository {
	return &sqliteGroupRepo{db: db}
}

func (r *sqliteGroupRepo) InitTable() error {
	groupTable := `CREATE TABLE IF NOT EXISTS groups (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        color TEXT DEFAULT '#64748b'
    );`

	junctionTable := `CREATE TABLE IF NOT EXISTS tablet_groups (
        tablet_id INTEGER NOT NULL,
        group_id INTEGER NOT NULL,
        PRIMARY KEY (tablet_id, group_id),
        FOREIGN KEY (tablet_id) REFERENCES tablets(id) ON DELETE CASCADE,
        FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
    );`

	if _, err := r.db.Exec(groupTable); err != nil {
		return err
	}
	_, err := r.db.Exec(junctionTable)
	return err
}

// Create utilise maintenant le struct pour passer description et color
func (r *sqliteGroupRepo) Create(g *Group) (int64, error) {
	query := `INSERT INTO groups (name, description, color) VALUES (:name, :description, :color)`
	res, err := r.db.NamedExec(query, g)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *sqliteGroupRepo) GetAll() ([]Group, error) {
	var groups []Group
	err := r.db.Select(&groups, "SELECT * FROM groups ORDER BY name ASC")
	return groups, err
}

func (r *sqliteGroupRepo) GetByID(id int64) (*Group, error) {
	var g Group
	err := r.db.Get(&g, "SELECT * FROM groups WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *sqliteGroupRepo) Update(g *Group) error {
	query := `UPDATE groups SET name=:name, description=:description, color=:color WHERE id=:id`
	_, err := r.db.NamedExec(query, g)
	return err
}

func (r *sqliteGroupRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM groups WHERE id = ?", id)
	return err
}

func (r *sqliteGroupRepo) AddTabletToGroup(tabletID, groupID int64) error {
	query := `INSERT OR IGNORE INTO tablet_groups (tablet_id, group_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, tabletID, groupID)
	return err
}

func (r *sqliteGroupRepo) RemoveTabletFromGroup(tabletID, groupID int64) error {
	query := `DELETE FROM tablet_groups WHERE tablet_id = ? AND group_id = ?`
	_, err := r.db.Exec(query, tabletID, groupID)
	return err
}

func (r *sqliteGroupRepo) GetGroupsByTablet(tabletID int64) ([]Group, error) {
	var groups []Group
	query := `
        SELECT g.* FROM groups g
        JOIN tablet_groups tg ON g.id = tg.group_id
        WHERE tg.tablet_id = ?`
	err := r.db.Select(&groups, query, tabletID)
	return groups, err
}

// Récupère toutes les tablettes appartenant à un groupe spécifique
func (r *sqliteGroupRepo) GetTabletsByGroup(groupID int64) ([]Tablet, error) {
	var tablets []Tablet
	query := `
        SELECT t.* FROM tablets t
        JOIN tablet_groups tg ON t.id = tg.tablet_id
        WHERE tg.group_id = ?`
	err := r.db.Select(&tablets, query, groupID)
	return tablets, err
}
