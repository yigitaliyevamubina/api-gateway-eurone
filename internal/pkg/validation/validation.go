package validation

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type customerValidation struct {
	Validate *validator.Validate
}

func NewCustomerValidation(validate *validator.Validate) *customerValidation {
	return &customerValidation{Validate: validate}
}

func (v *customerValidation) PhoneUz(fl validator.FieldLevel) bool {
	// get value 
	phone := strings.TrimSpace(fl.Field().String())
	// parse our phone number 
	isMatch, err := regexp.MatchString("^[9]{1}[9]{1}[8]{1}(?:77|88|93|94|90|91|95|93|99|97|98|33)[0-9]{7}$", phone)
	if err != nil {
		return false
	}
	return isMatch
}