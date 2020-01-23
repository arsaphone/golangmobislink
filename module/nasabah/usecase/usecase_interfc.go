package usecase

import (
	"github.com/gwlkm_service/module/nasabah/model"
	"database/sql"
)

type UseCaseNasabah interface {
	RegisterNasabah(input model.InputNasabah)(int,string,string)
	CheckLoginTelp(no_telp string)(int,string,string,string,string)
	CheckPin(nasabah_id,pin string)(int,string,string,string,string,string)
	GetTab(nasabah_id string)(error,[]*model.TabNasabah)
	GetSaldoPayment(nasabah_id string)(string,sql.NullFloat64,model.InfoNasabah,error)
	GantiPin(pin_lama,pin,pin_konfirmasi,nasabah_id string)(string,string,string)
	GetPoin(nasabah_id string)(string,string,string,model.NasabahPoin)
	RegisterAgen(nasabah_id string)(string,string,string,string)
	TransferKeSesamaLembaga(model.InputTransfer)(string,string,string,model.ResultTransfer)
	DetailRek(nasabah_id string,no_rekening string) model.DetailTabungan
	CheckEreg()(int,int)
	GetSaldoPay(nasabah_id,pin string)(string,string,string,string,int64)
	ERegistrasi(no_hp,nasabah_id string)(string,string,string)
	GetInfoNasabahByRekening(rekening string)(string,string,string,string,string)
	GetListTransaksi(no_rekening string,limit,offset int)([]*model.HistoryTrans,int,int)
}





