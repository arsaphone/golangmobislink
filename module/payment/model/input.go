package model

type InputPay struct{
	Tujuan         string    `form:"tujuan"           validate:"required"`
	NasabahID      string    `form:"nasabah_id"       validate:"required"`
	Nominal        int       `form:"nominal"          validate:"required"`
    Adm            int       `form:"adm"              validate:"required"`
	Reffid         string    `form:"reffid"           validate:"required"`
	Keterangan     string    `form:"keterangan"       validate:"required"`
}
