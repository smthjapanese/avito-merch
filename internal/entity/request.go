package entity

// AuthRequest запрос на аутентификацию
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SendCoinRequest запрос на отправку монет
type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}
