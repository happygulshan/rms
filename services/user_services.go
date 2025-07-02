package services

import (
	"database/sql"
)

func GetUserRoles(db *sql.DB, id string) []string {
	//Fetch role names for the user
	query := `
		SELECT r.name 
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := db.Query(query, id)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var userRoles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil
		}
		userRoles = append(userRoles, role)
	}
	return userRoles

}
