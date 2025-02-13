package entity

// AuthResponse ответ на успешную аутентификацию
type AuthResponse struct {
	Token string `json:"token"`
}

// InfoResponse ответ с информацией о пользователе
type InfoResponse struct {
	Coins       int64           `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

// InventoryItem элемент инвентаря
type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

// CoinHistory история транзакций с монетами
type CoinHistory struct {
	Received []CoinReceived `json:"received"`
	Sent     []CoinSent     `json:"sent"`
}

// CoinReceived информация о полученных монетах
type CoinReceived struct {
	FromUser string `json:"fromUser"`
	Amount   int64  `json:"amount"`
}

// CoinSent информация об отправленных монетах
type CoinSent struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Errors string `json:"errors"`
}
