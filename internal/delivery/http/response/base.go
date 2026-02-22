package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaseResponse struct {
	Status  int         `json:"status" example:"200"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func JSON(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, BaseResponse{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string, errs interface{}) {
	c.JSON(status, BaseResponse{
		Status:  status,
		Message: message,
		Errors:  errs,
	})
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, "Success", data)
}

func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, "Created successfully", data)
}
