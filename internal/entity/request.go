package entity

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int64  `json:"amount"`
}
