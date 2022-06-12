package model

type Wallet struct {
	User    int     `json: "user"`
	Balance float32 `json: "balance"`
}

type Transaction struct {
	Messages string  `json: "status"`
	Price    float32 `json: "price"`
}

type TopUp struct {
	Amount float32 `json: "amount"`
}

type ErrorMessage struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
