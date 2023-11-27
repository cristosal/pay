# Pay
Pay is a package that contains methods and functions related to accessing customer, plan, subscription, and price data. It was originally designed as part of a SAAS project and was extrapolated into a separate package.

Pay synchronizes with 3rd party vendors such as [stripe](https://www.stripe.com) and paypal, allowing you to backup and copy all of the data that exists on those platforms locally in your database. 

All entities are updated and kept in sync via  a `Webhook() http.Handler` interface that is exposed. 

Data that comes in through the webhook is treated as the source of truth, with the exception of adding customers all data  is to be added and updated from the payment provider. `pay` simply listens to these events and updates your entities accordingly. This is to avoid the recursive case where an update to an entity from the application triggers a webhook message which then updates the same entity, which then trigers a message to the webhook etc...

Currently `pay` only supports postgres as an underlying database. If you wish to support other databases please make a pull request.

## Installation

Same as any other go package
`go get -u github.com/cristosal/pay`

## Getting Started

In order to use pay you will need to create an `*sql.DB` object. Here I am using pgx as the driver

```go
db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
```

Then you can create the instance of pay service depending on which 3rd party provider you are using

```go
service := pay.NewStripeService(&pay.StripeConfig{
	Repo:          pay.NewEntityRepo(db),
	EventRepo:     pay.NewStripeEventRepo(db),
	Key:           os.Getenv("STRIPE_API_KEY"),
	WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
})
```

Initialize migrations and repository

```go
err := service.Init(context.TODO())
```

After this you can register the `http` webhook listener which will then update your entities upon recieving messages from the provider

```go
http.HandleFunc("/webhook/stripe", service.Webhook())
```
Pay is a package that contains methods and functions related to accessing customer, plan, subscription, and price data. It was originally designed as part of a SAAS project and was extrapolated into a separate package.

Pay synchronizes with 3rd party vendors such as [stripe](https://www.stripe.com) and paypal, allowing you to backup and copy all of the data that exists on those platforms locally in your database. 

All entities are updated and kept in sync via  a `Webhook() http.Handler` interface that is exposed. 

Data that comes in through the webhook is treated as the source of truth, with the exception of adding customers all data  is to be added and updated from the payment provider. `pay` simply listens to these events and updates your entities accordingly. This is to avoid the recursive case where an update to an entity from the application triggers a webhook message which then updates the same entity, which then trigers a message to the webhook etc...

Currently `pay` only supports postgres as an underlying database. If you wish to support other databases please make a pull request.

## Installation

Same as any other go package
`go get -u github.com/cristosal/pay`

## Getting Started

In order to use pay you will need to create an `*sql.DB` object. Here I am using pgx as the driver

```go
db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
```

Then you can create the instance of pay service depending on which 3rd party provider you are using

```go
service := pay.NewStripeService(&pay.StripeConfig{
	Repo:          pay.NewEntityRepo(db),
	EventRepo:     pay.NewStripeEventRepo(db),
	Key:           os.Getenv("STRIPE_API_KEY"),
	WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
})
```

Initialize migrations and repository

```go
err := service.Init(context.TODO())
```

After this you can register the `http` webhook listener which will then update your entities upon recieving messages from the provider

```go
http.HandleFunc("/webhook/stripe", service.Webhook())
```
