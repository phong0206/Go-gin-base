package helper

type IdDto struct {
	ID int `json:"id" binding:"required,min=1"`
}