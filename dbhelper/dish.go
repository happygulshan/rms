package dbhelper

import (
	"database/sql"
	"rms/models"
)

func GetAllDishes(db *sql.DB, limit int, offset int, restaurantID string) ([]models.Dish, error) {

	rows, err := db.Query("SELECT id, restaurant_id, name, description, price, created_at, created_by FROM dishes WHERE restaurant_id = $1::uuid ORDER BY created_at DESC LIMIT $2 OFFSET $3", restaurantID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var dishes []models.Dish

	for rows.Next() {
		var dish models.Dish
		if err := rows.Scan(&dish.ID, &dish.RestaurantID, &dish.Name, &dish.Description, &dish.Price, &dish.CreatedAt, &dish.CreatedBy); err != nil {
			return nil, err
		}
		dishes = append(dishes, dish)
	}

	return dishes, nil
}

func CreateDish(db *sql.DB, dish *models.Dish) error {
	err := db.QueryRow("INSERT INTO dishes (restaurant_id, Name, description, price, created_by) VALUES($1::uuid, $2, $3, $4, $5::uuid) RETURNING id, created_at",
		dish.RestaurantID, dish.Name, dish.Description, dish.Price, dish.CreatedBy).Scan(&dish.ID, &dish.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}
