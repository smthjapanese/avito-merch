package entity

type AuthResponse struct {
	Token string `json:"token"`
}

type InfoResponse struct {
	Coins       int64           `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

type CoinHistory struct {
	Received []CoinReceived `json:"received"`
	Sent     []CoinSent     `json:"sent"`
}

type CoinReceived struct {
	FromUser string `json:"fromUser"`
	Amount   int64  `json:"amount"`
}

type CoinSent struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}
