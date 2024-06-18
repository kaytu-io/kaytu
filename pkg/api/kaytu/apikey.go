package kaytu

import (
	"time"
)

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyResponse struct {
	ID        uint      `json:"id" example:"1"`                               // Unique identifier for the key
	Name      string    `json:"name" example:"example"`                       // Name of the key
	Active    bool      `json:"active" example:"true"`                        // Activity state of the key
	CreatedAt time.Time `json:"createdAt" example:"2023-03-31T09:36:09.855Z"` // Creation timestamp in UTC
	Token     string    `json:"token"`                                        // Token of the key
}
