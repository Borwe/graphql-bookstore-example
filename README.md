# Example implementation of a Book store using graphql

## Requirements
- Docker
- Go

Docker is required for running a postgres database instance, as passwords, username are hardcoded (this was just an example)

## Executing
- Build DB
```sh
docker build -t xyzdb ./
```
- Starting DB
```sh
docker run -p 5432:5432 -rm xyzdb
```
Note, name `xyzdb` to whatever you prefer

- Setting up env
```sh
go mod tidy
```

- Running server
```sh
go run main.go
```

Note, server uses port 8080 when it starts, and will print out queries it recieves, 
also errors should the occur due to bad input.

## Graphql schema implemented.

```graphql
type Book {
    id: ID!
    title: String!
    author: String!
    publishedYear: Int!
}

type Query {
    books: [Book!]!
    book(id: ID!): Book
}

type Mutation {
    createBook(input: BookInput!): Book!
    updateBook(id: ID!, input: BookInput!): Book!
    deleteBook(id: ID!): Boolean!
}

input BookInput {
    title: String!
    author: String!
    publishedYear: Int!
}
```

### Doing queries
I used curl

A simple createBook mutation might look like this:
```sh
curl localhost:8080/graphql -v -d 'mutation {createBook(input: {title:"Coot", author:"Boo", publishedYear:2026}){title author publishedYear}}'
```

An example of a book query to see authors would look like this.
```sh
curl localhost:8080/graphql -d 'query {books {id author}}'
```

As noted the endpoint is the `localhost:8080/graphql`
