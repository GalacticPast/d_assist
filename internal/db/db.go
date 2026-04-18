package db

import "fmt"

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func Check_if_user_exists(user_data User) bool {
	fmt.Println("ID: ", user_data.Id)
	fmt.Println("Email: ", user_data.Email)
	fmt.Println("Name: ", user_data.Name)
	return true
}
