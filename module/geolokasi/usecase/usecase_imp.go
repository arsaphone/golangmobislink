package usecase


import(
	"github.com/gwlkm_service/module/geolokasi/model"
	"github.com/gwlkm_service/module/geolokasi/repository"
)


type GeolokasiUseCaseImp struct{
   GeolokasiRepo  repository.GeolokasiRepo
}

func NewGeolokasiUsecaseImp(GeolokasiRepo repository.GeolokasiRepo)*GeolokasiUseCaseImp{
	return &GeolokasiUseCaseImp{
	  GeolokasiRepo : GeolokasiRepo,
	}
}

func(geo *GeolokasiUseCaseImp)LoadGeolokasi()(string,string,[]*model.OutputLokasi){
	list:=geo.GeolokasiRepo.LoadTempatNasabah()
	return "1","Success",list
}

func(geo *GeolokasiUseCaseImp)UpdateGeolokasi(input model.InputLokasi)(string,string,string){
	err:=geo.GeolokasiRepo.CheckNasabah(input.Nasabah_ID)
	if(err!=nil){
		err=geo.GeolokasiRepo.InsertTempatNasabah(input)
		if(err!=nil){
			return "0","Gagal Update Geolokasi",err.Error()
		}

		return "1","Berhasil Update Geolokasi",""
	}

	err=geo.GeolokasiRepo.UpdateTempatNasabah(input)
	if(err!=nil){
		return "0","Gagal Update Geolokasi",err.Error()
	}



	return "1","Berhasil Update Geolokasi",""
}
