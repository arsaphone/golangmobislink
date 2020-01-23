package usecase

import(
	"github.com/gwlkm_service/module/geolokasi/model"
)

type GeolokasiUseCase interface{
	LoadGeolokasi()(string,string,[]*model.OutputLokasi)
	UpdateGeolokasi(model.InputLokasi)(string,string,string)
}