package dbhelper

import (
	"database/sql"
	"rms/models"
)

func IsEmailTaken(db *sql.DB, email string) (bool, error) {
	var id string
	err := db.QueryRow("SELECT id FROM users WHERE email = $1 AND archived_at IS NULL", email).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func CreateUserWithRole(tx *sql.Tx, user *models.Users, hashedPassword, role, created_by string) error {

	var creator interface{} = nil
	if created_by != "" {
		creator = created_by
	}

	err := tx.QueryRow("INSERT INTO users(name, email, password, created_by) VALUES($1, $2, $3, $4) RETURNING id",
		user.Name, user.Email, hashedPassword, creator).Scan(&user.Id)
	if err != nil {
		return err
	}

	var roleID string
	err = tx.QueryRow("SELECT id FROM roles WHERE name = $1", role).Scan(&roleID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO user_roles(user_id, role_id) VALUES($1, $2)", user.Id, roleID)
	return err
}

func GetUserDetails(db *sql.DB, user *models.Users) error {
	// fetching detail of given user
	err := db.QueryRow("SELECT id, name, email, password FROM users WHERE email=$1 AND archived_at IS NULL", user.Email).
		Scan(&user.Id, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return err
	}
	return nil
}

func GetUserID(db *sql.DB, addID string) (string, error) {
	var userIDAdd string
	err := db.QueryRow("SELECT user_id FROM addresses WHERE id = $1", addID).Scan(&userIDAdd)

	if err != nil {
		return "", err
	}
	return userIDAdd, nil
}

func GetLatAndLngUser(db *sql.DB, addID string)(float64, float64, error) {

	var addLat, addLng float64
	err := db.QueryRow("SELECT lat, lng FROM addresses WHERE id = $1", addID).Scan(&addLat, &addLng)
	if err != nil {
		return 0, 0, err
	}
	return addLat, addLng, nil
}

func CreateAddress(db *sql.DB, add *models.Addresses) error {
	err := db.QueryRow("INSERT INTO addresses (user_id, lat, lng) VALUES($1::uuid, $2, $3) RETURNING id",
		add.UserID, add.Lat, add.Lng).Scan(&add.Id)

	if err != nil {
		return err
	}
	return nil
}
