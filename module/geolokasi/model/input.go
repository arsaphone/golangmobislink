package model

type InputLokasi struct{
    KodeLKM        string        `form:"kode_lkm"       validate:"required"`
	Nasabah_ID     string        `form:"nasabah_id"     validate:"required"`
	Nama_Nasabah   string        `form:"nama_nasabah"   validate:"required"`
	Lat            string        `form:"lat"            validate:"required"`
	Lng            string        `form:"lng"            validate:"required"`
	Keterangan     string        `form:"keterangan"     validate:"required"`
}


