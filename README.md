# Pay

Pay is a package that contains methods and functions related to accessing customer, plan, subscription, and price data.

Pay synchronizes with  [stripe](https://www.stripe.com)  allowing you to keep a local copy all of the data that exists on the platform in your database. 

All entities are updated and kept in sync via  a `Webhook() http.Handler` interface that is exposed by the service.

To see an example of a micro service that uses this package check out https://github.com/cristosal/micropay

## Installation

`go get -u github.com/cristosal/pay`

## Getting Started

Below is the example of how to get started using the `stripe` provider
 
Open an sql database
```go
db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
```

Create the provider
```go
provider := pay.NewStripeProvider(&pay.StripeConfig{
	Repo:          pay.NewEntityRepo(db),
	Key:           os.Getenv("STRIPE_API_KEY"),
	WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
})
```

Initialize tables and run migrations
```go
err := provider.Init(context.TODO())
```

Sync (pull) entities from the provider. In our case stripe
```go
err := provider.Sync()
```

Register the webhook which will update your entities upon receiving messages from the provider
```go
http.HandleFunc("/webhook/stripe", service.Webhook())
```

You have full access to querying the entities via the repository.

For instance you can list all plans available for purchase

```go
plans, err := provider.ListAllPlans()
```

**PLEASE NOTE:**
All `Add`, `Remove` and `Update` methods, do not go directly to the database. 
Instead, they issue a request to `stripe` to do the modification. Stripe will send an event to the webhook once the modification was complete. Upon receiving the webhook, the entities in the database will be updated. This ensures that your data is always in sync with stripe, and allows you to modify data from within the stripe dashboard itself.

In practice, doing a `ListAllCustomers` right after an `AddCustomer` is not guaranteed to return the customer that was added. You have to wait until stripe acknowledges the request and send data to your webhook.
## Events

You can hook into events susing any of the various `On` methods.

```go
provider.OnSubscriptionCreated(func (s *Subscription) {
	log.Printf("subscription with id %s created", s.ProviderID)
})

provider.OnSubscriptionUpdated(func (prev, current *Subscription) {
	if prev.Active != s.Active {
		log.Printf("subscription %s status changed", current.ProviderID)
	}
})

provider.OnSubscriptionRemoved(func (s *Subscription) {
	log.Printf("subscription %s deleted", s.ProviderID)
})
```

These events are available for Plans, Customers, and Prices as well.

