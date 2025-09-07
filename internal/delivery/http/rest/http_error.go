package rest

import (
	"errors"
	appErr "loopi-api/internal/common/errors"
	"net/http"
)

func HandleError(w http.ResponseWriter, err error) {
	var domainErr appErr.DomainError
	if errors.As(err, &domainErr) {
		WriteError(w, domainErr.Status(), domainErr.Message())
		return
	}
	ServerError(w, err.Error())
}
