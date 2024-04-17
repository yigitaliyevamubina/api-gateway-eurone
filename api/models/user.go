package models

type User struct {
	Id           string `json:"id" example:"uuid"`
	Username     string `json:"username" example:"john_doe"`
	Email        string `json:"email" example:"john@example.com"`
	Password     string `json:"password" example:"hashedpassword123"`
	FirstName    string `json:"firstname" example:"John"`
	LastName     string `json:"lastname" example:"Doe"`
	Bio          string `json:"bio" example:"Software engineer passionate about technology."`
	Website      string `json:"website" example:"https://johndoe.com"`
	IsActive     bool   `json:"is_active" example:"true"`
	RefreshToken string `json:"refresh_token" example:"somerandomrefresh123"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type GetListFilter struct {
	Page    int64  `json:"page"`
	Limit   int64  `json:"limit"`
	OrderBy string `json:"order_by"`
}

type Users struct {
	Count int64   `json:"count"`
	Users []*User `json:"users"`
}

type Message struct {
	Message string `json:"message"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}