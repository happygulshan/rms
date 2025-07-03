package dbhelper

import (
	"database/sql"
	"fmt"
	"log"
	"rms/models"
)

func GetAllRestaurants(db *sql.DB, limit int, offset int) ([]models.Restaurant, error) {
	rows, err := db.Query("SELECT id, name, description, lat, lng, created_at, created_by FROM restaurants ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	defer rows.Close()

	var restaurants []models.Restaurant

	for rows.Next() {
		var restaurant models.Restaurant
		if err := rows.Scan(&restaurant.ID, &restaurant.Name, &restaurant.Description, &restaurant.Lat, &restaurant.Lng, &restaurant.CreatedAt, &restaurant.CreatedBy); err != nil {
			log.Println(err.Error())
			return nil, err
		}

		restaurants = append(restaurants, restaurant)
	}
	return restaurants, nil

}

func CreateRestaurant(db *sql.DB, restaurant *models.Restaurant) error {
	err := db.QueryRow("INSERT INTO restaurants (Name, description, lat, lng, created_by) VALUES($1, $2, $3, $4, $5) RETURNING id",
		restaurant.Name, restaurant.Description, restaurant.Lat, restaurant.Lng, restaurant.CreatedBy).Scan(&restaurant.ID)

	if err != nil {
		return err
	}
	return nil
}

func RestaurantCreatedBy(db *sql.DB, resID string) (string, error) {
	var createdBy string
	err := db.QueryRow("SELECT created_by FROM restaurants WHERE id = $1::uuid", resID).Scan(&createdBy)
	if err != nil {
		return "", err
	}

	return createdBy, nil
}

func GetLatAndLngRest(db *sql.DB, resID string) (float64, float64, error) {

	var lat, lng float64
	err := db.QueryRow("SELECT lat, lng FROM restaurants WHERE id = $1", resID).Scan(&lat, &lng)
	if err != nil {
		return 0, 0, err
	}
	return lat, lng, nil
}
