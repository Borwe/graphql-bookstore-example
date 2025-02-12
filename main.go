package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
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

    queryFields := graphql.Fields{
	"books": &graphql.Field{
	    Type: graphql.String,
	    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return "YEAHHHHHH", nil
	    },
	},
    }

    rootQuery := graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: queryFields,
    }
    schemaConfig := graphql.SchemaConfig{
	Query: graphql.NewObject(rootQuery),
    }

    schema, err := graphql.NewSchema(schemaConfig)
    if err != nil {
	panic(fmt.Sprintf("Failed creating config with error %s\n", err))
    }

    http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
	var params struct {
	    Query string
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err!=nil {
	    http.Error(w, "Invalid request", http.StatusBadRequest)
	    return
	}

	result := graphql.Do(graphql.Params{
	    Schema: schema,
	    RequestString: params.Query,
	})

	if result.HasErrors() {
	    http.NotFound(w,r)
	    return
	}

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(result)
    })

    http.ListenAndServe(":8080",nil)
}
