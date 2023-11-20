module github.com/cristosal/pay

go 1.21.4

require (
	github.com/cristosal/dbx v1.0.0
	github.com/cristosal/migra v1.0.0
	github.com/cristosal/pgxx v1.0.0
	github.com/jackc/pgx/v5 v5.5.0
	github.com/stripe/stripe-go/v74 v74.30.0
)

replace github.com/cristosal/dbx => ../dbx

replace github.com/cristosal/migra => ../migra

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.15.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
