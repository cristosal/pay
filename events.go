package pay

type Events struct {
	subAddedCallbacks        []func(*Subscription)
	subUpdatedCallbacks      []func(*Subscription, *Subscription)
	subRemovedCallbacks      []func(*Subscription)
	customerAddedCallbacks   []func(*Customer)
	customerUpdatedCallbacks []func(*Customer, *Customer)
	customerRemovedCallbacks []func(*Customer)
	planAddedCallbacks       []func(*Plan)
	planUpdatedCallbacks     []func(*Plan, *Plan)
	planRemovedCallbacks     []func(*Plan)
	priceAddedCallbacks      []func(*Price)
	priceUpdatedCallbacks    []func(*Price, *Price)
	priceRemovedCallbacks    []func(*Price)
}

func (e *Events) OnSubscriptionAdded(cb func(*Subscription)) {
	e.subAddedCallbacks = append(e.subAddedCallbacks, cb)
}

func (e *Events) OnSubscriptionUpdated(cb func(*Subscription, *Subscription)) {
	e.subUpdatedCallbacks = append(e.subUpdatedCallbacks, cb)
}

func (e *Events) OnSubscriptionRemoved(cb func(*Subscription)) {
	e.subRemovedCallbacks = append(e.subRemovedCallbacks, cb)
}

func (e *Events) OnCustomerAdded(cb func(*Customer)) {
	e.customerAddedCallbacks = append(e.customerAddedCallbacks, cb)
}

func (e *Events) OnCustomerUpdated(cb func(*Customer, *Customer)) {
	e.customerUpdatedCallbacks = append(e.customerUpdatedCallbacks, cb)
}

func (e *Events) OnCustomerRemoved(cb func(*Customer)) {
	e.customerRemovedCallbacks = append(e.customerRemovedCallbacks, cb)
}

func (e *Events) OnPlanAdded(cb func(*Plan)) {
	e.planAddedCallbacks = append(e.planAddedCallbacks, cb)
}

func (e *Events) OnPlanUpdated(cb func(*Plan, *Plan)) {
	e.planUpdatedCallbacks = append(e.planUpdatedCallbacks, cb)
}

func (e *Events) OnPlanRemoved(cb func(*Plan)) {
	e.planRemovedCallbacks = append(e.planRemovedCallbacks, cb)
}

func (e *Events) OnPriceAdded(cb func(*Price)) {
	e.priceAddedCallbacks = append(e.priceAddedCallbacks, cb)
}

func (e *Events) OnPriceUpdated(cb func(*Price, *Price)) {
	e.priceUpdatedCallbacks = append(e.priceUpdatedCallbacks, cb)
}

func (e *Events) OnPriceRemoved(cb func(*Price)) {
	e.priceRemovedCallbacks = append(e.priceRemovedCallbacks, cb)
}

func (e *Events) subAdded(s *Subscription) {
	for _, cb := range e.subAddedCallbacks {
		cb(s)
	}
}

func (e *Events) subUpdated(prev *Subscription, s *Subscription) {
	for _, cb := range e.subUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *Events) subRemoved(s *Subscription) {
	for _, cb := range e.subRemovedCallbacks {
		cb(s)
	}
}

func (e *Events) customerAdded(c *Customer) {
	for _, cb := range e.customerAddedCallbacks {
		cb(c)
	}
}

func (e *Events) customerUpdated(prev *Customer, s *Customer) {
	for _, cb := range e.customerUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *Events) customerRemoved(c *Customer) {
	for _, cb := range e.customerRemovedCallbacks {
		cb(c)
	}
}

func (e *Events) planAdded(c *Plan) {
	for _, cb := range e.planAddedCallbacks {
		cb(c)
	}
}

func (e *Events) planUpdated(prev *Plan, s *Plan) {
	for _, cb := range e.planUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *Events) planRemoved(c *Plan) {
	for _, cb := range e.planRemovedCallbacks {
		cb(c)
	}
}

func (e *Events) priceAdded(c *Price) {
	for _, cb := range e.priceAddedCallbacks {
		cb(c)
	}
}

func (e *Events) priceUpdated(prev *Price, s *Price) {
	for _, cb := range e.priceUpdatedCallbacks {
		cb(prev, s)
	}
}

func (e *Events) priceRemoved(c *Price) {
	for _, cb := range e.priceRemovedCallbacks {
		cb(c)
	}
}
