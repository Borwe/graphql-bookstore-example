package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Book struct {
    gorm.Model
    title string
    author string
    publishedYear string
}

func main(){
    dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
    DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err!= nil {
        panic("Failed to connect")
    }

    DB.AutoMigrate(&Book{})
}
