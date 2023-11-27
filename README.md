# Pay

Pay is a package that contains methods and functions related to accessing customer, plan, subscription, and price data.

Pay synchronizes with 3rd party vendors such as [stripe](https://www.stripe.com) and paypal, allowing you to keep a platform agnostic copy all of the data that exists on those platforms in your database. 

All entities are updated and kept in sync via  a `Webhook() http.Handler` interface that is exposed by each providers service.

*Currently `pay` only supports postgres as an underlying database. If you wish to support other databases please make a pull request.*
## Installation

`go get -u github.com/cristosal/pay`

## Getting Started

Below is the example of how to get started using the `stripe` provider
 
Open the database
```go
db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
```

Create the provider
```go
service := pay.NewStripeProvider(&pay.StripeConfig{
	Repo:          pay.NewEntityRepo(db),
	EventRepo:     pay.NewStripeEventRepo(db),
	Key:           os.Getenv("STRIPE_API_KEY"),
	WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
})
```

Initialize tables and run migrations
```go
err := service.Init(context.TODO())
```

Sync (pull) entities from the provider
```go
err := service.Sync()
```

Register the webhook which will update your entities upon receiving messages from the provider
```go
http.HandleFunc("/webhook/stripe", service.Webhook())
```

You have full access to querying the entities via the repository
## Notes

Data that comes in through the webhook should be treated as the single source of truth. 

With the exception of adding customers for purposes of checking out, all data should be updated from the payment provider and received via webhook. 

>This is to avoid the recursive case where an update to an entity from the application triggers a webhook message which then updates the same entity, which then triggers a message to the webhook etc...

