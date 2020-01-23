package controller

import (
	"github.com/gin-gonic/gin"
    "gopkg.in/go-playground/validator.v9"
	"net/http"
	"github.com/gwlkm_service/module/geolokasi/usecase"
	"github.com/gwlkm_service/module/geolokasi/model"
)


type GeolokasiController struct{
	UseCase usecase.GeolokasiUseCase
}


func(geo *GeolokasiController)LoadLokasi(c *gin.Context){
	success,message,list:=geo.UseCase.LoadGeolokasi()
	c.JSON(http.StatusOK, gin.H{
		"success"   :  success,
		"message"   :  message,
		"data"      :  list,
	})
}

func(geo *GeolokasiController)UpdateLokasi(c *gin.Context){
	var validate *validator.Validate
	var input model.InputLokasi
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
	 
   success,message,desc_error:=geo.UseCase.UpdateGeolokasi(input)
   c.JSON(http.StatusOK, gin.H{
	"success"   : success,
	"message"   : message,
	"desc_error": desc_error,
   })
}



