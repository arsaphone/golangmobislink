package repository

import(
	"github.com/gwlkm_service/module/geolokasi/model"
)

type GeolokasiRepo interface{
	CheckNasabah(nasabah_id string)(error)
	LoadTempatNasabah()[]*model.OutputLokasi
	UpdateTempatNasabah(model.InputLokasi)(error)
	InsertTempatNasabah(model.InputLokasi)(error)
}