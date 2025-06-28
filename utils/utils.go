package utils

import (
	"math"
)

var RolePriorityMap = map[string]int{
	"user":       1,
	"subadmin":   2,
	"admin":      3,
	"superadmin": 4,
}

func HasOnlySubadminPrivileges(roles []string) bool {
	for _, role := range roles {
		if role == "admin" || role == "superadmin" {
			return false // They have higher privilege, so skip subadmin restrictions
		}
	}

	// No admin or superadmin roles found
	for _, role := range roles {
		if role == "subadmin" {
			return true
		}
	}
	return false
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func GetUserPriority(userRoles []string) int {
	var userPriority int = -1
	for _, role := range userRoles {
		userPriority = MaxInt(RolePriorityMap[role], userPriority)
	}
	return userPriority
}

// HaversineDistance calculates the distance between two lat/lng points in km
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth radius in kilometers

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
