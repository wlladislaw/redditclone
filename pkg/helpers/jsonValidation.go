package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
)

func UnmarshalAndValidate(r *http.Request, in interface{}) error {
	errDecode := json.NewDecoder(r.Body).Decode(in)
	defer r.Body.Close()
	if errDecode != nil {
		return errDecode
	}

	_, errValidation := govalidator.ValidateStruct(in)
	if errValidation != nil {
		return errValidation
	}
	return nil
}
