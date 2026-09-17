package health

// Dependencies interface defines the dependencies that are critical or optional for the transport layer to function correctly.
type Dependencies interface {
	Critical() []Reporter
	Optional() []Reporter
}
