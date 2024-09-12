package sub

// SubStore implements [parser.Parser]'s subscriptions interface
type SubStore struct {
	subscribers map[string]bool
}

// New creates a new SubscriptionManager
func New() *SubStore {
	return &SubStore{
		subscribers: make(map[string]bool),
	}
}

// Subscribe adds an address to the subscription list
func (sm *SubStore) Add(address string) bool {
	if sm.Check(address) {
		return false // Address already subscribed
	}
	sm.subscribers[address] = true
	return true
}

// Remove removes an address from the subscription list
func (sm *SubStore) Remove(address string) bool {
 if !sm.Check(address) {
  return false // Address not subscribed
 }
 delete(sm.subscribers, address)
 return true
}

// Check verifies if an address is subscribed
func (sm *SubStore) Check(address string) bool {
	_, exists := sm.subscribers[address]
	return exists
}
