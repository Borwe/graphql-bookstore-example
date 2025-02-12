package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/graphql-go/graphql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Book struct {
    Id uint `gorm:"primaryKey"`
    Title string
    Author string
    PublishedYear int
}

type BookInput struct {
    Title string
    Author string
    PublishedYear int
}

func main(){
    dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
    DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err!= nil {
        panic("Failed to connect")
    }

    if err:=DB.AutoMigrate(&Book{}); err!= nil {
	panic(fmt.Sprintf("DB migrate failed with error: %s\n", err))
    }

    //book type
    bookType := graphql.NewObject(
	graphql.ObjectConfig{
	    Name: "Book",
	    Fields: graphql.Fields{
		"id": &graphql.Field{
		    Type: graphql.NewNonNull(graphql.Int),
		},
		"title": &graphql.Field{
		    Type: graphql.NewNonNull(graphql.String),
		},
		"author": &graphql.Field{
		    Type: graphql.NewNonNull(graphql.String),
		},
		"publishedYear": &graphql.Field{
		    Type: graphql.NewNonNull(graphql.Int),
		},
	    },
	},
    )

    //book input type
    bookInputType := graphql.NewInputObject(
	graphql.InputObjectConfig{
	    Name: "BookInput",
	    Fields: graphql.InputObjectConfigFieldMap{
		"title": &graphql.InputObjectFieldConfig{
		    Type: graphql.NewNonNull(graphql.String),
		},
		"author": &graphql.InputObjectFieldConfig{
		    Type: graphql.NewNonNull(graphql.String),
		},
		"publishedYear": &graphql.InputObjectFieldConfig{
		    Type: graphql.NewNonNull(graphql.Int),
		},
	    },
	},
    )

    queryFields := graphql.Fields{
	"books": &graphql.Field{
	    Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(bookType))),
	    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		var books []Book
		DB.Find(&books)
		return books, nil
	    },
	},
	"book": &graphql.Field{
	    Type: bookType,
	    Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
		    Type: graphql.Int,
		},
	    },
	    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		v, ok := p.Args["id"].(int)
		if !ok {
		    return nil, nil
		}
		//get from db
		var book Book
		result := DB.Find(&book, v)
		if result.Error != nil {
		    return nil, result.Error
		}
		return book, nil
	    },
	},
    }


    mutationFields := graphql.Fields{
	"deleteBook": &graphql.Field{
	    Type: graphql.NewNonNull(graphql.Boolean),
	    Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
		    Type: graphql.NewNonNull(graphql.Int),
		},
	    },
	    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		id, ok := p.Args["id"].(int)
		if !ok {
		    return nil, errors.New("Need to pass `id` field")
		}
		result := DB.Delete(&Book{}, id)
		if result.Error != nil {
		    return nil, result.Error
		}
		return true, nil
	    },
	},
	"updateBook": &graphql.Field{
	    Type: graphql.NewNonNull(bookType),
	    Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
		    Type: graphql.Int,
		},
		"input": &graphql.ArgumentConfig{
		    Type: bookInputType,
		},
	    },
	    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		id, ok := p.Args["id"].(int)
		if !ok {
		    return nil, errors.New("No `id` field passed")
		}
		v, ok := p.Args["input"].(map[string]interface{})
		if !ok {
		    return nil, errors.New("No `input` field passed")
		}

		var book Book
		result := DB.Find(&book, &id)
		if result.Error != nil {
		    return nil, errors.New("Book with for updating not found")
		}
		book.Author = v["author"].(string)
		book.PublishedYear = v["publishedYear"].(int)
		book.Title = v["title"].(string)
		fmt.Println("COME ON!!!!", book)
		result = DB.Save(&book)
		if result.Error != nil {
		    return nil, result.Error
		}
		return book, nil
	    },
	},
	"createBook": &graphql.Field{
	    Type: graphql.NewNonNull(bookType),
	    Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
		    Type: bookInputType,
		},
	    },
	    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		v, ok := p.Args["input"].(map[string]interface{})
		if !ok{
		    return nil, errors.New("Not valid input passed")
		}

		book := Book{
		    Title: v["title"].(string),
		    Author: v["author"].(string),
		    PublishedYear: v["publishedYear"].(int),
		}

		result := DB.Create(&book)
		if result.Error != nil {
		    return nil, result.Error
		}
		return book, nil
	    },
	},
    }

    rootQuery := graphql.ObjectConfig{
	Name: "Query",
	Fields: queryFields,
    }

    mutations := graphql.ObjectConfig{
	Name: "Mutations",
	Fields: mutationFields,
    }
    schemaConfig := graphql.SchemaConfig{
	Query: graphql.NewObject(rootQuery),
	Mutation: graphql.NewObject(mutations),
    }

    schema, err := graphql.NewSchema(schemaConfig)
    if err != nil {
	panic(fmt.Sprintf("Failed creating config with error %s\n", err))
    }

    http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
	rawquery, err :=  io.ReadAll(r.Body)
	if err!=nil {
	    http.Error(w, "Invalid requestsdsadasd", http.StatusBadRequest)
	    return
	}

	query := string(rawquery)

	fmt.Println("Q IS:",query)

	result := graphql.Do(graphql.Params{
	    Schema: schema,
	    RequestString: query,
	})

	if result.HasErrors() {
	    fmt.Println("HAS ERROS:")
	    fmt.Println(result.Errors)
	    w.WriteHeader(http.StatusBadRequest)
	    w.Write([]byte(fmt.Sprintln(result.Errors)))
	    return
	}

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(result)
    })

    http.ListenAndServe(":8080",nil)
}
