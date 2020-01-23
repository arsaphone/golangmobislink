package model

type InputNasabah struct{
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

type InputTransfer struct{
	Tujuan         string    `form:"tujuan"           validate:"required"`
	NasabahID      string    `form:"nasabah_id"       validate:"required"`
	Nominal        int       `form:"nominal"          validate:"required"`
	Pin            string    `form:"pin"              validate:"required"`
}