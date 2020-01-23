package controller

import (
	"github.com/gin-gonic/gin"
    "gopkg.in/go-playground/validator.v9"
	_"github.com/gwlkm_service/config"
	"net/http"
	"github.com/gwlkm_service/module/payment/usecase"
	"github.com/gwlkm_service/module/payment/model"
)


type PaymentController struct{
	 UseCase usecase.PaymentUseCase 
}

func(payment *PaymentController)PayHandler(c *gin.Context){
	var validate *validator.Validate
	var inputPay model.InputPay
	validate = validator.New()
	c.Bind(&inputPay)
	errors:=validate.Struct(inputPay)
	if errors != nil {
	  c.JSON(http.StatusOK, gin.H{
			"success"   :  0,
			"message"   : "Kesalahan Server",
			"desc_error":  errors.Error(),
		})
	  return
   }
 
	status,message,desc:=payment.UseCase.PayHandler(inputPay)
		c.JSON(http.StatusOK, gin.H{
		  "success"      :  status,
		  "message"      :  message,
		  "desc_error"   :  desc,
		})
}



