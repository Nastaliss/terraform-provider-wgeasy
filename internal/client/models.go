package client

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Client represents a WireGuard client/peer as returned by the wg-easy API.
type Client struct {
	ID                  FlexibleID `json:"id"`
	UserID              int64      `json:"userId"`
	InterfaceID         string     `json:"interfaceId"`
	Name                string     `json:"name"`
	Enabled             bool       `json:"enabled"`
	IPv4Address         string     `json:"ipv4Address"`
	IPv6Address         string     `json:"ipv6Address"`
	PublicKey           string     `json:"publicKey"`
	PrivateKey          string     `json:"privateKey"`
	PresharedKey        string     `json:"preSharedKey"`
	ExpiresAt           *string    `json:"expiresAt"`
	AllowedIPs          []string   `json:"allowedIps"`
	ServerAllowedIPs    []string   `json:"serverAllowedIps"`
	DNS                 []string   `json:"dns"`
	MTU                 int64      `json:"mtu"`
	PersistentKeepalive int64      `json:"persistentKeepalive"`
	ServerEndpoint      *string    `json:"serverEndpoint"`
	PreUp               string     `json:"preUp"`
	PostUp              string     `json:"postUp"`
	PreDown             string     `json:"preDown"`
	PostDown            string     `json:"postDown"`
	JC                  int64      `json:"jC"`
	JMin                int64      `json:"jMin"`
	JMax                int64      `json:"jMax"`
	I1                  *string    `json:"i1"`
	I2                  *string    `json:"i2"`
	I3                  *string    `json:"i3"`
	I4                  *string    `json:"i4"`
	I5                  *string    `json:"i5"`
	OneTimeLink         *string    `json:"oneTimeLink"`
	CreatedAt           string     `json:"createdAt"`
	UpdatedAt           string     `json:"updatedAt"`
}

// FlexibleID handles JSON values that may be a string or a number,
// normalizing them to a string.
type FlexibleID string

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	// Try string first.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleID(s)
		return nil
	}
	// Try number.
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleID(strconv.FormatInt(int64(n), 10))
		return nil
	}
	return fmt.Errorf("clientId is neither string nor number: %s", string(data))
}

func (f FlexibleID) String() string {
	return string(f)
}

// CreateClientRequest is the body for POST /api/client.
type CreateClientRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expiresAt"`
}

// CreateClientResponse is the response from POST /api/client.
type CreateClientResponse struct {
	Status   string     `json:"status"`
	ClientID FlexibleID `json:"clientId"`
}

// UpdateClientRequest is the body for POST /api/client/:id.
// ALL fields are required by the API. Nullable fields use pointers (nil -> JSON null).
type UpdateClientRequest struct {
	// Required non-nullable fields
	Name                string   `json:"name"`
	Enabled             bool     `json:"enabled"`
	IPv4Address         string   `json:"ipv4Address"`
	IPv6Address         string   `json:"ipv6Address"`
	ServerAllowedIPs    []string `json:"serverAllowedIps"`
	MTU                 int64    `json:"mtu"`
	PersistentKeepalive int64    `json:"persistentKeepalive"`
	PreUp               string   `json:"preUp"`
	PostUp              string   `json:"postUp"`
	PreDown             string   `json:"preDown"`
	PostDown            string   `json:"postDown"`
	JC                  int64    `json:"jC"`
	JMin                int64    `json:"jMin"`
	JMax                int64    `json:"jMax"`
	// Nullable fields - nil serializes to JSON null
	ExpiresAt      *string  `json:"expiresAt"`
	AllowedIPs     []string `json:"allowedIps"`
	DNS            []string `json:"dns"`
	ServerEndpoint *string  `json:"serverEndpoint"`
	I1             *string  `json:"i1"`
	I2             *string  `json:"i2"`
	I3             *string  `json:"i3"`
	I4             *string  `json:"i4"`
	I5             *string  `json:"i5"`
}
