package handler

import (
	"books-and-trust/services/api-gateway/util"
	"net/http"

)

// HeathCheck godoc
//
//	@Summary		Server Health Check
//	@Description	This method outputs the current status and metrics of the API Gateway
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string	"everything is fine"
//	@Failure		500	{object}	string	"Internal Server Error"
//	@Router			/health [get]
func (h *HTTPHandler) HeathCheck(w http.ResponseWriter, r *http.Request) {
	if err := util.WriteJSON(w, "server metrics", http.StatusOK, "everything is fine"); err != nil {
		util.InternalServerErr(w, r, h.Logger, err)
	}
}
