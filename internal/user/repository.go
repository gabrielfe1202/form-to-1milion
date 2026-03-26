package user

import "database/sql"

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(u User) error {
	query := `INSERT INTO users (name, email, document, phone) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Exec(query, u.Name, u.Email, u.Document, u.Phone)
	return err
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	var u User
	err := r.DB.QueryRow(`SELECT id, name, email, document, phone FROM users WHERE email = $1`, email).Scan(&u.ID, &u.Name, &u.Email, &u.Document, &u.Phone)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (r *Repository) GetByDocument(document string) (*User, error) {
	var u User
	err := r.DB.QueryRow(`SELECT id, name, email, document, phone FROM users WHERE document = $1`, document).Scan(&u.ID, &u.Name, &u.Email, &u.Document, &u.Phone)
	if err != nil {
		return nil, nil
	}
	return &u, nil
}

func (r *Repository) List() ([]User, error) {
	rows, err := r.DB.Query(`SELECT id, name, email, document, phone FROM users`)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Document, &u.Phone); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) Count() (int, error) {
	var totalCount int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalCount)
	if err != nil {
		return 0, err
	}
	return totalCount, nil
}
