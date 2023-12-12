
# Pay

Pay serves as an intermediary between your application and third party payment providers such as  [stripe](https://www.stripe.com), allowing you to query all plan, customer and subscription data without directly going to the provider.  It also allows you to create checkout sessions so that your customers can purchase your plans.

Pay synchronizes with the provider and stores data in a way that is provider-agnostic. 
Data can be kept in sync via  a `Webhook() http.Handler`  that receives events from the provider, or manually via the `Sync` method. This approach boosts your application's robustness, speeds up data retrieval, and allows you to support multiple providers.

To see an example of a micro service that uses this package check out https://github.com/cristosal/cent

## Installation

`go get -u github.com/cristosal/pay`

## Usage

Below is the example of how to use the `stripe` provider. Note that error handling has been omitted for brevity.

### Initialization
 
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

Initialize the provider. This will create payment tables by running the migrations
```go
err := provider.Init(context.TODO())
```

### Syncing Data

To have a local copy of data that exists in our stripe account, it is good practice to run the `Sync` method any time the application starts so as to always have up-to-date data.

```go
// Adds or updates our database with the newest data available
err := provider.Sync()
```

### Receive updates from the provider

In order to keep the data in sync during the lifetime of our application we need to receive updates from the provider to our `Webhook`.

```go
// Register the webhook
http.HandleFunc("/webhook/stripe", provider.Webhook())

// Start the http server
http.ListenAndServe(":8080", nil)
```

## Checkout

When our customers want to purchase a plan at a specific pricing we can give them a url to visit to checkout. 

**IMPORTANT:** All `Add`, `Remove` and `Update` methods in the `provider` object, do not go directly to the database. Instead, they issue a request to the provider (Stripe) to perform the action. Stripe will then send a POST request to the `webhook`. Upon receiving the message,  entities in the database will be updated. In practice, doing a `ListAllCustomers` right after an `AddCustomer` is not guaranteed to return the customer that was added. You have to wait until stripe acknowledges the request and sends the event to your `webhook`. This approach ensures that your data is always in sync with stripe, and allows changes within the stripe dashboard itself to be reflected in your database.
### Add Plan

Note that you do not need to specify the `Provider` as we have already created the provider as a `StripeProvider` so this field will get populated automatically. Both `ID` and `ProviderID` don't exist yet so those fields are blank as well.

```go
err := provider.AddPlan(&pay.Plan{
	Name:        "Basic Plan",
	Description: "Access all basic features",
	Active:      true,
})
```

### Add Price

After the plan is saved we will add a price to it

```go
err := provider.AddPrice(&pay.Price{
	PlanID:    1,    // replace with your plan id
	Amount:    1000, // this is in cents. The equivalent would be $10.00
	Currency: "USD",
	Schedule: PricingMonthly,
})
```


### Add Customer

Next let's add the customer

```go
err := provider.AddCustomer(&pay.Customer{
	Name: "Test Customer",
	Email: "test@example.com",
})
```

### Redirect to checkout

With all our entities in place we can now perform the checkout. The `RedirectURL` property is the url a user will redirect to once they have completed the checkout

```go
url, err := provider.Checkout(&pay.CheckoutRequest{
	CustomerID:  1,    // id of our customer
	PriceID:     1,    // id of our price attached to a plan
	RedirectURL: "http://myapp.com/success",
})
```

The `url` return variable contains the url that a user can go to actually perform the checkout. If you are using `pay` within the context of a web app you can redirect the user as follows

```go
func HandleCheckout(w http.ResponseWriter, r *http.Request) {
	// ... your checkout logic here
	
	// assuming everything is correct and we have url variable available
	http.Redirect(w, r, url, http.StatusSeeOther)
}
```
## Events

You can hook into events using any of the various `On` methods.

```go
provider.OnSubscriptionCreated(func (s *Subscription) {
	log.Printf("subscription with id %s created", s.ProviderID)
})

provider.OnSubscriptionUpdated(func (prev, current *Subscription) {
	if prev.Active != s.Active {
		log.Printf("subscription %s status changed to %v", 
			current.ProviderID, current.Active)
	}
})

provider.OnSubscriptionRemoved(func (s *Subscription) {
	log.Printf("subscription %s deleted", s.ProviderID)
})
```

These events are available for Plans, Customers, and Prices as well.

## Associating multiple users with a subscription

Sometimes you want to associate multiple users with one subscription. This can be the case in seat-based plans where a customer can give `x` amount of users access to an account.

>When a subscription is first added, a `SubscriptionUser` is added with username being the email of the customer that purchased the subscription.

To help facilitate this we have the `SubscriptionUser` entity

```go
type SubscriptionUser struct {
	SubscriptionID int64
	Username       string
}
```

*Username is the unique identifier for the user.*

We can manage seats by using the following methods

```go
func (r *Repo) AddSubscriptionUser(su *SubscriptionUser) error

func (r *Repo) RemoveSubscriptionUser(su *SubscriptionUser) error

func (r *Repo) CountSubscriptionUsers(subID int64) (int64, error)
```

If we want to get the underlying subscription or plan for the user...

```go
func (r *Repo) GetSubscriptionByUsername(username string) (*Subscription, error)

func (r *Repo) GetPlanByUsername(username string) (*Plan, error)
```
