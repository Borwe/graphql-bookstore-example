# Example implementation of a Book store using graphql

## Requirements
- Docker
- Go

Docker is required for running a postgres database instance, as passwords, username are hardcoded (this was just an example)

## Executing
```sh
go run main.go
```

Not, server uses port 8080 when it starts, and will print out queries it recieves, 
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

