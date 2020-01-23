package model

type InputRegisterViaAgen struct{
	NasabahID       string        `form:"nasabah_id"    validate:"required"`
	Nama            string        `form:"nama"          validate:"required"`
	Alamat          string        `form:"alamat"        validate:"required"`
	Email           string        `form:"email"         validate:"required"`
	NIK             string        `form:"nik"           validate:"required"`
	NoHP            string        `form:"nohp"          validate:"required"`
	TempatLahir     string        `form:"tempat_lahir"  validate:"required"`
	Tgl             string        `form:"tgl"           validate:"required"`
	Bulan           string        `form:"bulan"         validate:"required"`
	Tahun           string        `form:"tahun"         validate:"required"`
	JK              string        `form:"jk"            validate:"required"`
}


type InputTab struct{
	IDNasabah      string        `form:"nasabah_id"            validate:"required"`
	KodeProduk     string        `form:"kode_produk"           validate:"required"`
	IDUser         string        `form:"nasabah_id_agen"       validate:"required"`
}