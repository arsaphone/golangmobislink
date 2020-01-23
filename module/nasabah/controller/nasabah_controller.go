package controller

import (
  _"github.com/gwlkm_service/module/nasabah/repository"
	"github.com/gin-gonic/gin"
  "gopkg.in/go-playground/validator.v9"
  _"github.com/gwlkm_service/config"
	"net/http"
	"github.com/gwlkm_service/module/nasabah/usecase"
	"github.com/gwlkm_service/module/nasabah/model"
	"github.com/gwlkm_service/helpers"
	jwt "github.com/dgrijalva/jwt-go"
	"strconv"
	"github.com/dustin/go-humanize"


)


type NasabahController struct{
	 UseCase usecase.UseCaseNasabah 
}



func(nasabah *NasabahController)RegisterNasabah(c *gin.Context){
  var validate *validator.Validate
  var inputNasabah model.InputNasabah
  validate = validator.New()
  c.Bind(&inputNasabah)
  errors:=validate.Struct(inputNasabah)
  if errors != nil {
    c.JSON(http.StatusOK, gin.H{
          "success"   : 0,
		  "message"   : "Kesalahan Server",
		  "desc_error":errors.Error(),
      })
    return
 }
	
 success,message,desc_error:=nasabah.UseCase.RegisterNasabah(inputNasabah)

 c.JSON(http.StatusOK, gin.H{
	 "success"   :  success,
	 "message"   :  message,
	 "desc_error":  desc_error,
 })

}

func(nasabah *NasabahController)GetTabNasabah(c *gin.Context){
   type Tabungan struct {
	Nasabah_id       string                      `json:"NASABAH_ID"           db:"NASABAH_ID"`
	NoRekening       string                      `json:"NO_REKENING"          db:"NO_REKENING"`
    KodeProduk       string                      `json:"KODE_PRODUK"          db:"KODE_PRODUK"`
    DeskripsiProduk  string                      `json:"DESKRIPSI_PRODUK"     db:"DESKRIPSI_PRODUK"`
    SaldoAkhir       string                     `json:"SALDO_AKHIR"          db:"SALDO_AKHIR"`
   }

   var Listtab []Tabungan
   var tab Tabungan
	nasabah_id:=c.PostForm("nasabah_id")
	err,list:=nasabah.UseCase.GetTab(nasabah_id)
	if(err!=nil){
		c.JSON(http.StatusOK, gin.H{
			"success"   :  0,
		  "message"   :  "Anda tidak punya tabungan..",
		  "desc_error":  err.Error(),
		})
		return
	}

    for _,v :=range list {

		tab=Tabungan{
			Nasabah_id      : v.Nasabah_id,
			NoRekening      : v.NoRekening,
			KodeProduk      : v.KodeProduk,
			DeskripsiProduk : v.DeskripsiProduk,
			SaldoAkhir      : humanize.Comma(int64(v.SaldoAkhir.Float64)),
		}
		Listtab=append(Listtab,tab)
	}

  c.JSON(http.StatusOK,Listtab)
}

func(nasabah *NasabahController)Transfer(c *gin.Context){
	var validate *validator.Validate
	var inputTransfer model.InputTransfer
	validate = validator.New()
	c.Bind(&inputTransfer)
	errors:=validate.Struct(inputTransfer)
	if errors != nil {
	  c.JSON(http.StatusOK, gin.H{
			"success"   :  0,
			"message"   : "Kesalahan Server",
			"desc_error":  errors.Error(),
		})
	  return
   }
 
	status,message,desc,result:=nasabah.UseCase.TransferKeSesamaLembaga(inputTransfer)
		c.JSON(http.StatusOK, gin.H{
		  "success"                 :  status,
		  "message"                 :  message,
		  "desc_error"              :  desc,
		  "no_rekening_pengirim"    :  result.Pengirim,
		  "nama_pengirim"           :  result.NamaPengirim,
		  "no_rekening_penerima"    :  result.Penerima,
		  "nama_penerima"           :  result.NamaPenerima,
		  "waktu"                   :  result.Waktu,
		  "nominal"                 :  result.Nominal,
		})
}



func(nasabah *NasabahController)SaldoNasabah(c *gin.Context){
	id:=c.PostForm("nasabah_id")
	no_rekening,saldo_akhir,info,err:=nasabah.UseCase.GetSaldoPayment(id)
	if(err!=nil){
		c.JSON(http.StatusOK, gin.H{
			"success"   :  0,
		  "message"   :  "Anda tidak punya tabungan Sukarela..",
		  "desc_error":  err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success"       :  1,
		"message"       :  "Berhasil",
		"desc_error"    :  "",
		"no_rekening"   : no_rekening,
		"saldo_sukarela":"Rp. "+humanize.Comma(int64(saldo_akhir.Float64)),
		"nama_nasabah"  : info.NamaNasabah,
		"email"         : info.Email,
		"tgl_lahir"     : info.TanggalLahir,
		"tgl_register"  : info.TglRegister,
		"tempat_lahir"  : info.TempatLahir,
		"jenis_kelamin" : info.JenisKelamin,
		"point"         : info.TotalPoint,
	})

}

func(nasabah *NasabahController)CheckEreg(c *gin.Context){
	jumlah_nasabah,jumlah_akun:=nasabah.UseCase.CheckEreg()
	c.JSON(http.StatusOK,gin.H{
			"success"        : "1",
			"message"        : "Selamat Datang",
			"jumlah_nasabah" : jumlah_nasabah,
			"nasabah_aktif"  : jumlah_akun,
			"nasabah_non"    : jumlah_nasabah-jumlah_akun,
	})
}

func(nasabah *NasabahController)Ereg(c *gin.Context){
	no_hp:=c.PostForm("no_hp")
	nasabah_id:=c.PostForm("nasabah_id")
	success,message,desc:=nasabah.UseCase.ERegistrasi(no_hp,nasabah_id)
	c.JSON(http.StatusOK,gin.H{
			"success"        : success,
			"message"        : message,
			"desc_error"     : desc,
	})
}


func(nasabah *NasabahController)GetSaldoPay(c *gin.Context){
	nasabah_id:=c.PostForm("id")
	pin:=c.PostForm("pin")
	success,message,email,norek,jumlah:=nasabah.UseCase.GetSaldoPay(nasabah_id,pin)
	c.JSON(http.StatusOK,gin.H{
			"success"        : success,
			"message"        : message,
			"email"          : email,
			"no_rekening"    : norek,
			"saldo"          : jumlah,
	})
}


func(nasabah *NasabahController) DetailRek(c *gin.Context){
  norek:=c.PostForm("norek")
	idnasabah:=c.PostForm("nasabah_id")
	result:=nasabah.UseCase.DetailRek(idnasabah,norek)
	c.JSON(http.StatusOK, gin.H{
		"success"          :  "1",
		"message"          :  "Selamat Datang",
		"nasabah_id"        :  result.Nasabah_id,
		"no_rekening"            :  result.No_rekening,
		"deskripsi_produk" :  result.Deskripsi_produk,
		"saldo"            :  result.Saldo,
	})
}


func(nasabah *NasabahController)LoginNasabah(c *gin.Context){
	type_request:=c.Query("type")
	if(type_request=="1024"){
		 phone:=c.PostForm("no_hp")
		 status,msg,desc,nasabah_id,no_telp:=nasabah.UseCase.CheckLoginTelp(phone)
		 if(status==0){
				c.JSON(http.StatusOK, gin.H{
					"success"   :  status,
					"message"   :  msg,
					"desc_error":  desc,
				})
				return
		 }

		 c.JSON(http.StatusOK, gin.H{
			"success"   :  status,
			"message"   :  msg,
			"desc_error":  desc,
			"nasabah_id": nasabah_id,
			"no_hp"     : no_telp,
		})
	}

  if(type_request=="1025"){
		nasabah_id:=c.PostForm("nasabah_id")
		pin:=c.PostForm("pin")
		status,msg,desc,nasabah_id,no_telp,status_agen:=nasabah.UseCase.CheckPin(nasabah_id,pin)
		if(status==0){
			c.JSON(http.StatusOK, gin.H{
				"success"   :  status,
				"message"   :  msg,
				"desc_error":  desc,
			})
			return
		}

    sign := jwt.New(jwt.GetSigningMethod("HS256"))
	  token, err := sign.SignedString([]byte("secret_sbk"))
	  if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{
			  "success"   :  status,
			  "message"   :  "Gagal menyiapkan token..",
		   	"desc_error": err.Error(),
		 })
		 return
	  }


		c.JSON(http.StatusOK, gin.H{
			"success"          :  status,
			"message"          :  msg,
			"desc_error"       :  desc,
			"nasabah_id"       :  nasabah_id,
			"no_hp"            :  no_telp,
			"status_agen"      :  status_agen,
			"token"            :  token,
		})

	}
}


func(nasabah *NasabahController)GantiPinNasabah(c *gin.Context){
	nasabah_id:=helpers.Sha1(c.PostForm("nasabah_id"))
	pin_lama:=helpers.Sha1(c.PostForm("pin"))
	pin_baru:=c.PostForm("pin_baru")
	konfirmasi_pin:=c.PostForm("konfirmasi_pin")
	status,message,desc:=nasabah.UseCase.GantiPin(pin_lama,pin_baru,konfirmasi_pin,nasabah_id)

	c.JSON(http.StatusOK, gin.H{
		"success"   :  status,
		"message"   :  message,
		"desc_error":  desc,
	})
}

func(nasabah *NasabahController)GetPoin(c *gin.Context){
	nasabah_id:=c.PostForm("nasabah_id")
	status,message,desc,nasabah_x:=nasabah.UseCase.GetPoin(nasabah_id)

	c.JSON(http.StatusOK, gin.H{
		"success"   :  status,
		"message"   :  message,
		"desc_error":  desc,
		"nasabah_id":  nasabah_x.Nasabah_id,
		"nama"      :  nasabah_x.NamaNasabah,
		"point"     :  nasabah_x.TotalPoin,
		"no_hp"     :  nasabah_x.Telpon,
	})
}

func(nasabah *NasabahController)RegisterAgen(c *gin.Context){
	nasabah_id:=c.PostForm("nasabah_id")
	status,message,desc,nasabah_x:=nasabah.UseCase.RegisterAgen(nasabah_id)
	c.JSON(http.StatusOK, gin.H{
		"success"   :  status,
		"message"   :  message,
		"desc_error":  desc,
		"agen_id"   :  nasabah_x,
	})
}

func(nasabah *NasabahController)GetInfoNasabahByRekening(c *gin.Context){
	no_rekening:=c.PostForm("no_rekening")
	status,message,nasabah_id,nama_nasabah,no_rekenings:=nasabah.UseCase.GetInfoNasabahByRekening(no_rekening)
	c.JSON(http.StatusOK, gin.H{
		"success"        :  status,
		"message"        :  message,
		"nasabah_id"     :  nasabah_id,
		"nama_nasabah"   :  nama_nasabah,
		"no_rekening"    :  no_rekenings,
	})
}

func(nasabah *NasabahController)GetListTransaksi(c *gin.Context){
	no_rekening:=c.PostForm("no_rekening")
	halaman_sekarang,_:=strconv.Atoi(c.Param("page"))
	offset:=(halaman_sekarang-1)*10

    list,jumlah,jumlah_halaman:=nasabah.UseCase.GetListTransaksi(no_rekening,10,offset)
	c.JSON(http.StatusOK, gin.H{
		"success"        :  "1",
		"message"        :  "berhasil",
		"data"           :  list,
		"jumlah_data"    :  jumlah,
		"jumlah_halaman" :  jumlah_halaman,
	})
}



