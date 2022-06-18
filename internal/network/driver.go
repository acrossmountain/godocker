package network

type NetworkDriver interface {
	// Name return driver name
	Name() string

	// Create create network
	Create(subnet string, name string) (*Network, error)

	// Delete delete network
	Delete(network Network) error

	// Connect connect container network to endpoint network
	Connect(network *Network, endpoint *Endpoint) error

	// Disconnect
	Disconnect(network *Network, endpoint *Endpoint) error
}
