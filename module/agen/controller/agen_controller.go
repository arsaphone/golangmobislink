package controller

import (
	"github.com/gin-gonic/gin"
    "gopkg.in/go-playground/validator.v9"
	"net/http"
	"github.com/gwlkm_service/module/agen/usecase"
	"github.com/gwlkm_service/module/agen/model"
)

type AgenController struct{
	UseCase usecase.UseCaseAgen
}

func(agen *AgenController)RegisterViaAgen(c *gin.Context){
	var validate *validator.Validate
	var input model.InputRegisterViaAgen
	validate = validator.New()
	c.Bind(&input)
	errors:=validate.Struct(input)
	if errors != nil {
	  c.JSON(http.StatusOK, gin.H{
			"success"   : 0,
			"message"   : "Kesalahan Server",
			"desc_error":errors.Error(),
		})
	  return
   }
	  
   success,message,desc_error:=agen.UseCase.RegisterViaAgen(input)
  
   c.JSON(http.StatusOK, gin.H{
	   "success"   :  success,
	   "message"   :  message,
	   "desc_error":  desc_error,
   })
  }

  func(agen *AgenController)GetListAnggota(c *gin.Context){
	nasabah_id:=c.PostForm("nasabah_id")
	list_produk:=agen.UseCase.GetListAnggota(nasabah_id)
	c.JSON(http.StatusOK,list_produk)
  }

  func(agen *AgenController)GetTabProgram(c *gin.Context){
		list_produk:=agen.UseCase.GetTabProgram()
		c.JSON(http.StatusOK,list_produk)
	}

  func(agen *AgenController)CreateTabProgram(c *gin.Context){
		var validate *validator.Validate
		var input model.InputTab
		validate = validator.New()
		c.Bind(&input)
		errors:=validate.Struct(input)
		if errors != nil {
			c.JSON(http.StatusOK, gin.H{
				"success"   : 0,
				"message"   : "Kesalahan Server",
				"desc_error":errors.Error(),
			})
			return
		 }

		 success,message,desc_error:=agen.UseCase.CreateTabProgram(input)

		 c.JSON(http.StatusOK, gin.H{
			 "success"   :  success,
			 "message"   :  message,
			 "desc_error":  desc_error,
		 })
	}
