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

var FakeDB = []Book{
    Book{
	Id: 0,
	Title: "A",
	Author: "A",
	PublishedYear: 2024,
    },
    Book{
	Id: 1,
	Title: "B",
	Author: "B",
	PublishedYear: 2025,
    },
}

type Book struct {
    Id uint `gorm:"primarykey"`
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
		return FakeDB, nil
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
		return FakeDB[v], nil
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

		newBooks := []Book{}
		found := false
		for _, b := range FakeDB {
		    if b.Id == uint(id){
			found = true
			continue
		    }
		    newBooks = append(newBooks, b)
		}
		FakeDB = newBooks
		return found, nil
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
		newBooks := []Book{}
		for _, b := range FakeDB {
		    if b.Id == uint(id) {
			b.Author = v["author"].(string)
			b.PublishedYear = v["publishedYear"].(int)
			b.Title = v["title"].(string)
			book = b
			newBooks = append(newBooks, book)
			continue
		    }
		    newBooks = append(newBooks, b)
		}
		FakeDB = newBooks
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

		l := len(FakeDB)
		book := Book{
		    Id: uint(l),
		    Title: v["title"].(string),
		    Author: v["author"].(string),
		    PublishedYear: v["publishedYear"].(int),
		}

		FakeDB = append(FakeDB, book)
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
	    fmt.Println(result.Errors)
	    http.NotFound(w,r)
	    return
	}

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(result)
    })

    http.ListenAndServe(":8080",nil)
}
