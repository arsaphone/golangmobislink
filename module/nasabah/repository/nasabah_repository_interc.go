package repository

import (
	"github.com/gwlkm_service/module/nasabah/model"
	_"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)

type NasabahRepo interface {
	CheckNoTelp(no_telp string)error
	CheckNoNIK(no_nik string)error
	GetTemplateNID(nid string)string
	GetKodeKantor()string
	GetMaxNasabahId()string
	GetTabProduk() []*model.TabProduk	
	CheckNoTelponLogin(no_telp string)(string,string,error)
	CheckRegPayService(nasabah_id string,pin string)(error)
	CheckStatusAgenNasabah(nasabah_id string)(string)
	GetRegPayAccountByNasabahID(nasabah_id string)(string,string,error)
	GetTabNasabah(nasabah_id string)(error,[]*model.TabNasabah)
	GetOneTabProduk()model.TabProduk
	GetNasabahTabPayment(nasabah_id,kode_produk string)(string,sql.NullFloat64,error)
	GetNamaNasabah(nasabah_id string)string
	GetPinByNasabahId(nasabah_id string)(error,string) 
	UpdatePinNasabah(nasabah_id string,pin_baru string)error
	GetPoinByNasabahId(nasabah_id string)(error,model.NasabahPoin)
	UpdateStatusAgenNasabah(nasabah_id string)error
	GetTemplateRekening()string
	GetPayProductMapping(code string)model.PayProductMapping
	GetMaxTransID()string
	CheckValidNoRekeningNasabah(no_rek string)error
	CheckSaldoByNoRekening(no_rekening string)(string,sql.NullFloat64)
	GetDeskripsiTabProduk(kode_produk string)(string)
	GetKodeTabProduk(no_rekening string)(string)
	GetJumlahNasabah()(int)
	GetJumlahRegPay()(int)
	GetEmailNasabah(nasabah_id string)(string)
	CheckValidNasabahID(nasabah_id string)(error)
	GenerateNasabahID()(string)
	GetCountTransaksi(no_rekening string)(int)
	GetListTransaksi(no_rekening string,limit,offset int)([]*model.HistoryTrans)
	GetInfoNasabah(nasabah_id string)(model.InfoNasabah)
}
