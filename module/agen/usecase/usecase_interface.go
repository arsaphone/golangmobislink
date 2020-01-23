package usecase
import(
"github.com/gwlkm_service/module/agen/model"

)


type UseCaseAgen interface{
	RegisterViaAgen(input model.InputRegisterViaAgen)(string,string,string)
	GetTabProgram()[]*model.TabProduk
	CreateTabProgram(inputTab model.InputTab)(string,string,string)
	GetListAnggota(nasabah_id string)[]*model.ListAnggota

}