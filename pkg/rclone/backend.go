package rclone

// Backend is the low-level transport interface for rclone RPC calls.
// This allows using either embedded librclone or HTTP remote connections.
type Backend interface {
	// Call makes an rclone RC API call.
	// method is the RC method (e.g., "config/listremotes", "operations/list")
	// params is a JSON string of parameters
	// Returns the JSON response string and an HTTP-style status code (200 = success)
	Call(method string, params string) (string, int)
}
