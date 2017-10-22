package main

import "fmt"

func dbInit() string {
	return fmt.Sprintf("host=%s " +
		"port=%s " +
		"user=%s " +
		"password=%s " +
		"dbname=%s " +
		"sslmode=%s",
		C.DB.Host,
		C.DB.Port,
		C.DB.User,
		C.DB.Pass,
		C.DB.Name,
		C.DB.SSL,
	)
}
