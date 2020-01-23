package main

import (
  nasabahController  "github.com/gwlkm_service/module/nasabah/controller"
  nasabahRepository  "github.com/gwlkm_service/module/nasabah/repository"
  nasabahUseCase     "github.com/gwlkm_service/module/nasabah/usecase"
  agenController     "github.com/gwlkm_service/module/agen/controller"
  agenRepository     "github.com/gwlkm_service/module/agen/repository"
  agenUseCase        "github.com/gwlkm_service/module/agen/usecase"
  paymentUseCase     "github.com/gwlkm_service/module/payment/usecase"
  paymentController  "github.com/gwlkm_service/module/payment/controller"
  geolokasiController  "github.com/gwlkm_service/module/geolokasi/controller"
  geolokasiRepository  "github.com/gwlkm_service/module/geolokasi/repository"
  geolokasiUseCase     "github.com/gwlkm_service/module/geolokasi/usecase"
  jwt "github.com/dgrijalva/jwt-go"
  "github.com/gwlkm_service/config"
  "github.com/gin-gonic/gin"
  "net/http"
  "fmt"
  "github.com/BurntSushi/toml"
)


func main(){
	var toml_config config.TomlConfig
	if _, err := toml.DecodeFile("config.toml", &toml_config); err != nil {
		fmt.Println(err)
	}
	
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	db,err    := config.Connectdb(toml_config)
	dbsys,err := config.ConnectdbSys(toml_config)


	if(err!=nil){
       fmt.Println(err.Error())
	}

	nasabah_repository:=nasabahRepository.NewSQLNasabahRepo(db.DBE,dbsys.DB)
	nasabah_usecase:=nasabahUseCase.NewNasabahUsecaseImp(nasabah_repository,db.DBE)
	nasabah_controller:=nasabahController.NasabahController{UseCase : nasabah_usecase }

	agen_repository:=agenRepository.NewSQLAgenRepo(db.DBE,dbsys.DB)
	agen_usecase:=agenUseCase.NewAgenUsecaseImp(nasabah_repository,agen_repository,db.DBE)
	agen_controller:=agenController.AgenController{UseCase : agen_usecase}

	payment_usecase:=paymentUseCase.NewPaymentUsecaseImp(nasabah_repository,agen_repository,db.DBE)
	payment_controller:=paymentController.PaymentController{UseCase : payment_usecase}

	geolokasi_repository:=geolokasiRepository.NewSQLGeoRepo(db.DBE)
	geolokasi_usecase:=geolokasiUseCase.NewGeolokasiUsecaseImp(geolokasi_repository)
	geolokasi_controller:=geolokasiController.GeolokasiController{UseCase : geolokasi_usecase}

	v1 := r.Group("/nasabah")

	{
		v1.POST("/register",nasabah_controller.RegisterNasabah)
		v1.POST("/login", nasabah_controller.LoginNasabah)
		v1.POST("/tabungan",auth,nasabah_controller.GetTabNasabah)
		v1.POST("/saldo",auth,nasabah_controller.SaldoNasabah)
		v1.POST("/ganti_pin",auth,nasabah_controller.GantiPinNasabah)
		v1.POST("/poin",auth,nasabah_controller.GetPoin)
		v1.POST("/register_agen",auth, nasabah_controller.RegisterAgen)
		v1.POST("/transfer",auth, nasabah_controller.Transfer)
		v1.POST("/detail_rek",auth,nasabah_controller.DetailRek)
		v1.GET("/check_ereg",auth,nasabah_controller.CheckEreg)
		v1.POST("/saldo_pay",auth,nasabah_controller.GetSaldoPay)
		v1.POST("/e_registrasi",auth,nasabah_controller.Ereg)
		v1.POST("/info_rekening_nasabah",auth,nasabah_controller.GetInfoNasabahByRekening)
		v1.POST("/history_transaksi/halaman/:page",auth,nasabah_controller.GetListTransaksi)
	}

	v2:=r.Group("/agen")

	{
	   v2.POST("/register_via_agen",auth, agen_controller.RegisterViaAgen)
	   v2.GET("/get_tab_program",auth,agen_controller.GetTabProgram)
	   v2.POST("/create_tab_program",auth, agen_controller.CreateTabProgram)
	   v2.POST("/list_anggota",auth, agen_controller.GetListAnggota)

	}


	v3:=r.Group("/payment")
	{
		v3.POST("/pay_handler",auth,payment_controller.PayHandler)
	}


	v4:=r.Group("/geolokasi")
	{
		v4.GET("/load_lokasi",geolokasi_controller.LoadLokasi)
		v4.POST("/update_lokasi",geolokasi_controller.UpdateLokasi)
	}

	
	fmt.Println("Server running on port : "+toml_config.Server.Port)
	r.GET("/",root)
	r.Run(":"+toml_config.Server.Port)
}


func root(c *gin.Context){
	c.Header("Content-Type", "application/json") 
	c.JSON(http.StatusOK, gin.H{
		"message": "welcome my api",
	})
}

func auth(c *gin.Context) {
	tokenString := c.Request.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod("HS256") != token.Method {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		fmt.Println("Token GAgal")

		return []byte("secret_sbk"), nil
	})

	// if token.Valid && err == nil {
	if token != nil && err == nil {
		fmt.Println("token verified")
	} else {
		result := gin.H{
			"message": "not authorized",
			"error":   err.Error(),
		}
		c.JSON(http.StatusUnauthorized, result)
		c.Abort()
	}
}

