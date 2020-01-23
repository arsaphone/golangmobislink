package repository
import(
_ "database/sql"
"github.com/gwlkm_service/module/agen/model"
)


type AgenRepo interface{
	GetAgenKeyValue(keyvalue string) float64
	GetSumTabungan()float64
	GetSaldoAkhirPerkiraan(kode_perkiraan string) float64
	GetReferalID(nasabah_id string)(string,error)
	GetTabProgram()([]*model.TabProduk)
	GetBiayaTabProgram(kode_produk string)(string,float64,float64)
	CheckNasabahID(nasabah_id string)(error)
	GetCountTabungan(nasabah_id string)(int)
	GetBiayaAdminTabungan(keyname,kode_program string)(string)
	GetListAnggota(nasabah_id string)([]*model.ListAnggota)
	GetNoRekeningByJenisTabungan(nasabah_id,jenis string) string
	GetSetoranPertamaJenisTabungan(jenis_tabungan string) float64
	
}