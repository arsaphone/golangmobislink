package usecase

import (
	"github.com/gwlkm_service/module/nasabah/model"
	"github.com/gwlkm_service/config"
	"github.com/gwlkm_service/module/nasabah/repository"
	"github.com/gwlkm_service/helpers"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"github.com/jmoiron/sqlx"
	"time"
	"strings"
	"strconv"
	_"github.com/schigh/str"
	"database/sql"
	_"reflect"
	"github.com/dustin/go-humanize"
	_"github.com/kpango/glg"
	"log"
	"math"
)

type NasabahUseCaseImp struct{
	NasabahRepo repository.NasabahRepo
	DB *sqlx.DB
}


func NewNasabahUsecaseImp(NasabahRepo repository.NasabahRepo,DB *sqlx.DB)*NasabahUseCaseImp{
	return &NasabahUseCaseImp{
	  NasabahRepo : NasabahRepo,
	  DB          : DB,
	}
}

func(nasabah *NasabahUseCaseImp) GetSaldoPayment(nasabah_id string)(string,sql.NullFloat64,model.InfoNasabah,error){
	///nama_nasabah:=nasabah.NasabahRepo.GetNamaNasabah(nasabah_id)
	info:=nasabah.NasabahRepo.GetInfoNasabah(nasabah_id)
	produk:=nasabah.NasabahRepo.GetOneTabProduk()
	fmt.Println(produk.KodeProduk)
	no_rekening,saldo_akhir,err:=nasabah.NasabahRepo.GetNasabahTabPayment(nasabah_id,produk.KodeProduk)
	if(err!=nil){
		return "",saldo_akhir,info,err
	}
	return no_rekening,saldo_akhir,info,nil
}


func(nasabah *NasabahUseCaseImp)CheckLoginTelp(no_telp string) (int,string,string,string,string){
	nasabah_id,telp,err:=nasabah.NasabahRepo.CheckNoTelponLogin(no_telp)
	if(err!=nil){
		return 0,"No HP tidak terdaftar",err.Error(),"",""
	}
	return 1,"Berhasil","",nasabah_id,telp
}

func(nasabah *NasabahUseCaseImp)GetTab(nasabah_id string) (error,[]*model.TabNasabah){
	err,list_tab:=nasabah.NasabahRepo.GetTabNasabah(nasabah_id)
	if(err!=nil){
       return err,nil
	}
	return nil,list_tab
}

func(nasabah *NasabahUseCaseImp)DetailRek(nasabah_id string,no_rekening string) model.DetailTabungan{
	var result model.DetailTabungan
	nasabah_id2,saldo_penerima:=nasabah.NasabahRepo.CheckSaldoByNoRekening(no_rekening)
	kode_produk:=nasabah.NasabahRepo.GetKodeTabProduk(no_rekening)
	nama_nasabah:=nasabah.NasabahRepo.GetNamaNasabah(nasabah_id)
	deskripsi:=nasabah.NasabahRepo.GetDeskripsiTabProduk(kode_produk)
	result.Nasabah_id=nasabah_id2
	result.Deskripsi_produk=deskripsi
	result.Nama_nasabah=nama_nasabah
	result.No_rekening=no_rekening
	result.Saldo="Rp. "+humanize.Comma(int64(saldo_penerima.Float64))
	return result
}

func(nasabah *NasabahUseCaseImp)GetInfoNasabahByRekening(rekening string)(string,string,string,string,string){
    rekening_penerima:=""
	template_rekening:=nasabah.NasabahRepo.GetTemplateRekening()
	detect_rekening:=strings.Index(rekening,".")
	detect_template:=strings.Index(template_rekening,".")
	if(detect_rekening==detect_template){
		rekening_penerima=rekening
	} else {
		kode_kantor:=nasabah.NasabahRepo.GetKodeKantor()
		template_rekening:=nasabah.NasabahRepo.GetTemplateRekening()
        rekening_penerima=helpers.FixRekening(rekening,kode_kantor,template_rekening)
	}

	nasabah_id_penerima,_:=nasabah.NasabahRepo.CheckSaldoByNoRekening(rekening_penerima)
	nama_penerima:=nasabah.NasabahRepo.GetNamaNasabah(nasabah_id_penerima)
	err:=nasabah.NasabahRepo.CheckValidNoRekeningNasabah(rekening_penerima)
	if(err!=nil){
		return "0","No Rekening Nasabah tidak ditemukan","","",""
	}

	return "1","Berhasil",nasabah_id_penerima,nama_penerima,rekening_penerima
}


func(nasabah *NasabahUseCaseImp)TransferKeSesamaLembaga(input model.InputTransfer)(string,string,string,model.ResultTransfer){
	
	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	f, err := os.OpenFile(dir+"/error_transfer_lembaga.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	

	var result model.ResultTransfer
	currentTime := time.Now()
	tanggal_sekarang:=currentTime.Format("2006-01-02")
	timestamp := currentTime.Format("2006-01-02 15:04:05")  
	time := currentTime.Format("15:04:05")    
  
	pin:=input.Pin
	nasabah_id:=input.NasabahID

	err=nasabah.NasabahRepo.CheckRegPayService(helpers.Sha1(nasabah_id),helpers.Sha1(pin))
	if(err!=nil){
		return "0","Pin yang anda masukkan salah",err.Error(),result
	}
	if(pin=="123456"){
		return "0","Mohon harap ganti pin anda terlebih dahulu","",result
	}
	
	produk:=nasabah.NasabahRepo.GetOneTabProduk()
	no_rekening_pengirim,jumlah,err:=nasabah.NasabahRepo.GetNasabahTabPayment(nasabah_id,produk.KodeProduk)
	if(err!=nil){
		return "0","Anda tidak memiliki tabungan",err.Error(),result
	}
	nominal:=float64(input.Nominal)
	//saldo_rekening_pengirim_setelah_transfer:=jumlah.Float64-nominal
	if(nominal>jumlah.Float64){
		return "0","Saldo Anda tidak mencukupi","",result
	}
	rek_penerima:=""
	kode_kantor:=nasabah.NasabahRepo.GetKodeKantor()
	//len_kode_kantor:=len(kode_kantor)
	template_rekening:=nasabah.NasabahRepo.GetTemplateRekening()
	detect_rekening:=strings.Index(input.Tujuan,".")
	detect_template:=strings.Index(template_rekening,".")
	// fmt.Println(detect_rekening)
	// fmt.Println(detect_template)
	if(detect_rekening==detect_template){
		rek_penerima=input.Tujuan
	} else {

	}
	// fmt.Println(len_kode_kantor)
     fmt.Println("no_rek_penerima "+rek_penerima)
	 fmt.Println("no_rek_pengirim "+no_rekening_pengirim)
	// fmt.Println(saldo_rekening_pengirim_setelah_transfer)
	err=nasabah.NasabahRepo.CheckValidNoRekeningNasabah(rek_penerima)
	if(err!=nil){
		return "0","No Rekening Nasabah tidak ditemukan",err.Error(),result
	}
	 
	trans_id:=nasabah.NasabahRepo.GetMaxTransID()
	//trans_id:=nasabah.NasabahRepo.GenerateNasabahID()
	trans_id=helpers.CreateTransID(trans_id)
	//fmt.Println("Generated : "+trans_id_x)

	nasabah_id_penerima,_:=nasabah.NasabahRepo.CheckSaldoByNoRekening(rek_penerima)
	nama_pengirim:=nasabah.NasabahRepo.GetNamaNasabah(nasabah_id)
	//nama_penerima:=nasabah.NasabahRepo.GetNamaNasabah(nasabah_id_penerima)
	desc_kredit:=nasabah.NasabahRepo.GetPayProductMapping("TINTCR")
	descr_kredit:=desc_kredit.Deskripsi+" dari Rek "+no_rekening_pengirim+" a/n "+nama_pengirim+" Rp. "+humanize.Comma(int64(nominal))
	kwitansi_kredit:="TINTCR"+desc_kredit.Kode_Trans
	//saldo_penerima_setelah_transfer:=nominal+saldo_penerima.Float64
	// fmt.Println(descr_debit)
	// fmt.Println(nasabah_id_penerima)
	// fmt.Println(saldo_penerima.Float64)
	// fmt.Println(kwitansi_debit)
	nama_penerima:=nasabah.NasabahRepo.GetNamaNasabah(nasabah_id_penerima)
	desc_debit:=nasabah.NasabahRepo.GetPayProductMapping("TINTDB")
	descr_debit:=desc_debit.Deskripsi+" ke Rek "+rek_penerima+" a/n "+nama_penerima+" Rp. "+humanize.Comma(int64(nominal))
	kwitansi_debit:="TINTDB"+desc_debit.Kode_Trans
	// fmt.Println(descr_kredit)
	// fmt.Println(nasabah_id)
	// fmt.Println(kwitansi_kredit)

	tx,err_sql:= nasabah.DB.Begin()
	if(err_sql!=nil){
		fmt.Println(err_sql.Error())
	}

	fmt.Println("kode trabtrans_penerima : "+trans_id)

	if _, err_sql = tx.Exec("insert into TABTRANS(TABTRANS_ID,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans,issuerId,device_id)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id,tanggal_sekarang,rek_penerima,"316","100",nominal,kwitansi_kredit,
	"9990",descr_kredit,"1","T",no_rekening_pengirim,kode_kantor,time,timestamp,tanggal_sekarang,no_rekening_pengirim,"3213213123123"); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan transfer,terdapat kesalahan pada operasi insert tabtrans penerima,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	trans_id_int,_:=strconv.Atoi(trans_id)
	transaks_master_penerima:=trans_id_int+1

	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	transaks_master_penerima,"TAB",kwitansi_kredit,tanggal_sekarang,descr_kredit,"TAB","139128003",
	"9990","",kode_kantor,"0",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan transfer,terdapat kesalahan pada operasi insert transaksi_master penerima,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	 transaks_detail_penerima1:=transaks_master_penerima+1
	 transaks_detail_penerima2:=transaks_detail_penerima1+1

	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	transaks_detail_penerima1,transaks_master_penerima,"20101",nominal,"0","(NULL)",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi insert transaksi_detail debit penerima,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	transaks_detail_penerima2,transaks_master_penerima,"10505","0",nominal,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi insert transaksi_detail kredit penerima,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}


	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE NO_REKENING =? ",nominal,rek_penerima); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi update saldo penerima,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	tabtrans_pengirim:=transaks_detail_penerima1+1

	if _, err_sql = tx.Exec("insert into TABTRANS(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans,issuerId,device_id)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_pengirim,tanggal_sekarang,no_rekening_pengirim,
	"315","200",nominal,kwitansi_debit,"9990",descr_debit,"1","T",rek_penerima,kode_kantor,
	time,timestamp,tanggal_sekarang,rek_penerima,"3213213123123"); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi insert tabtrans pengirim,desc_error :"+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	tabtrans_pengirim_master:=tabtrans_pengirim+1

	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	tabtrans_pengirim_master,"TAB",kwitansi_debit,tanggal_sekarang,descr_debit,"TAB","139128003",
	"9990","",kode_kantor,"0",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi insert transaksi_master pengirim ,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}
	
	tabtrans_pengirim_detail1:=tabtrans_pengirim_master+1
	tabtrans_pengirim_detail2:=tabtrans_pengirim_detail1+1

	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	tabtrans_pengirim_detail1,tabtrans_pengirim_master,"20101",nominal,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi insert transaksi_detail debit pengirim,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	tabtrans_pengirim_detail2,tabtrans_pengirim_master,"10505","0",nominal,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi insert transaksi_detail kredit pengirim,desc_error : "+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR-? WHERE NO_REKENING =? ",nominal,no_rekening_pengirim); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada operasi update saldo pengirim,desc_error :"+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}

	err_sql=tx.Commit()
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal transfer,terdapat kesalahan pada saat commit,desc_error :"+err_sql.Error() )
		return "0","Kesalahan Server",err_sql.Error(),result
	}
	
	result.NamaPengirim=nama_pengirim
	result.NamaPenerima=nama_penerima
	result.Pengirim=no_rekening_pengirim
	result.Penerima=rek_penerima
	result.Nominal=nominal
	result.Waktu=timestamp

	return "1","Transfer Berhasil ","",result
}




func(nasabah *NasabahUseCaseImp)CheckPin(nasabah_id,pin string)(int,string,string,string,string,string){
	err:=nasabah.NasabahRepo.CheckRegPayService(helpers.Sha1(nasabah_id),helpers.Sha1(pin))
	if(err!=nil){
		return 0,"Pin anda salah",err.Error(),"","",""
	}
	nasabah_id,telp,err:=nasabah.NasabahRepo.GetRegPayAccountByNasabahID(nasabah_id)
	if(err!=nil){
		return 0,"Nasabah ID tidak ditemukan",err.Error(),"","",""
	}

	status:=nasabah.NasabahRepo.CheckStatusAgenNasabah(nasabah_id)
	if(status==""){
		return 1,"Berhasil","",nasabah_id,telp,"Belum Terdaftar"
	}

	return 1,"Berhasil","",nasabah_id,telp,"Terdaftar"
} 

func (nasabah *NasabahUseCaseImp) GantiPin(pin_lama,pin,pin_konfirmasi,nasabah_id string)(string,string,string){
	err,pin_lama_db:=nasabah.NasabahRepo.GetPinByNasabahId(nasabah_id)
	if(err!=nil){
		return "0","Nasabah ID tidak ditemukan",err.Error()
	}

    if(pin_lama==pin_lama_db){
		if(pin==pin_konfirmasi){
			pin_baru_x:=helpers.Sha1(pin)
			err=nasabah.NasabahRepo.UpdatePinNasabah(nasabah_id,pin_baru_x)
			if(err!=nil){
				return "0","Kesalahan Server Silakan Hubungi CS",err.Error()
			}
			return "1","Berhasil Mengganti Pin",""
		} else {
			return "0","Pin Baru dan Konfirmasi Pin Salah !",""
		}
	} else {
		return "0","Pin Lama Salah !",""
	}
}

func(nasabah *NasabahUseCaseImp)GetPoin(nasabah_id string)(string,string,string,model.NasabahPoin){
	err,nasabah_poin:=nasabah.NasabahRepo.GetPoinByNasabahId(nasabah_id)
	if(err!=nil){
		return "0","Nasabah ID tidak ada dalam Database",err.Error(),nasabah_poin
	}
	return "1","Berhasil","",nasabah_poin
}

func(nasabah *NasabahUseCaseImp) RegisterAgen(nasabah_id string)(string,string,string,string){
	produk:=nasabah.NasabahRepo.GetOneTabProduk()
	_,saldo_akhir,err:=nasabah.NasabahRepo.GetNasabahTabPayment(nasabah_id,produk.KodeProduk)
	if(err!=nil){
		return "0","Saldo Akhir Nasabah tidak ditemukan",err.Error(),""
	}

	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	

	var toml_config config.TomlConfig
	if _, err = toml.DecodeFile(dir+"/config.toml", &toml_config); err != nil {
		fmt.Println("Error terjadi "+ err.Error())
	}
	
	biaya_agen:=toml_config.Var_regagtbyagen.BIAYA_DAFTAR_AGEN
	if(saldo_akhir.Float64>biaya_agen){
		err=nasabah.NasabahRepo.UpdateStatusAgenNasabah(nasabah_id)
		if(err!=nil){
			return "0","gagal mendaftar jadi agen",err.Error(),""
		}
        return "1","Berhasil mendaftar jadi agen","",nasabah_id
	}
     return "0","Gagal mendaftar jadi agen karena saldo minimum tidak mencukupi","",nasabah_id
}

func(nasabah *NasabahUseCaseImp)CheckEreg()(int,int){
	jumlah_nasabah:=nasabah.NasabahRepo.GetJumlahNasabah()
	jumlah_akun:=nasabah.NasabahRepo.GetJumlahRegPay()
	return jumlah_nasabah,jumlah_akun
}

func(nasabah *NasabahUseCaseImp)GetListTransaksi(no_rekening string,limit,offset int)([]*model.HistoryTrans,int,int){
	list_x:=nasabah.NasabahRepo.GetListTransaksi(no_rekening,limit,offset)
	jumlah_transaksi:=nasabah.NasabahRepo.GetCountTransaksi(no_rekening)
	jumlah_halaman_x:=math.Ceil(float64(jumlah_transaksi)/float64(limit))
	return list_x,jumlah_transaksi,int(jumlah_halaman_x)
}


func(nasabah *NasabahUseCaseImp)GetSaldoPay(nasabah_id,pin string)(string,string,string,string,int64){
	err:=nasabah.NasabahRepo.CheckValidNasabahID(nasabah_id)

	if(err!=nil){
		return "0","Nasabah ID tidak terdaftar","","",0
	}

	err=nasabah.NasabahRepo.CheckRegPayService(helpers.Sha1(nasabah_id),helpers.Sha1(pin))
	if(err!=nil){
		return "0","Pin yang anda masukkan salah","","",0
	}
	if(pin=="123456"){
		return "0","Mohon harap ganti pin anda terlebih dahulu","","",0
	}

	produk:=nasabah.NasabahRepo.GetOneTabProduk()
	no_rekening_pengirim,jumlah,err:=nasabah.NasabahRepo.GetNasabahTabPayment(nasabah_id,produk.KodeProduk)
	if(err!=nil){
		return "0","Anda tidak punya rekening sukarela","","",0
	}

	email:=nasabah.NasabahRepo.GetEmailNasabah(nasabah_id)
	return "1","Berhasil",email,no_rekening_pengirim,int64(jumlah.Float64)
}


func(nasabah *NasabahUseCaseImp)ERegistrasi(no_hp,nasabah_id string)(string,string,string){
	err,_:=nasabah.NasabahRepo.GetPoinByNasabahId(nasabah_id)
	if(err!=nil){
		return "0","Nasabah tidak terdaftar",err.Error()
	}
	err=nasabah.NasabahRepo.CheckValidNasabahID(nasabah_id)
	if(err!=nil){
		encrypt_password:=helpers.Sha1("123456")
		encrypt_nasabah_id:=helpers.Sha1(nasabah_id)
		tx,err_sql:= nasabah.DB.Begin()

		_,err_sql=tx.Exec("INSERT INTO reg_pay_account VALUES (?,?,?,?,?,?)",nasabah_id,no_hp,"","","0","9991")
		if(err_sql!=nil){
			tx.Rollback()
			return "0","Kesalahan Server",err_sql.Error()
		}
	
		_,err_sql=tx.Exec("INSERT INTO register_pay_service VALUES (?,?,?,?,?,?,?)",
		encrypt_nasabah_id,"0017","",encrypt_password,"","4","1")
		if(err_sql!=nil){
			tx.Rollback()
			return "0","Kesalahan Server",err_sql.Error()
		}

		err_sql=tx.Commit()
		if(err_sql!=nil){
			tx.Rollback()
			return "0","Kesalahan Server",err_sql.Error()
		}
		
		return "1","E-Registrasi Sukses",""
	}
	
   return "0","Nasabah sudah terdaftar",""
}





func(nasabah *NasabahUseCaseImp)RegisterNasabah(input model.InputNasabah) (int,string,string){
    currentTime := time.Now()
	tanggal_register:=currentTime.Format("2006-01-02")
	tgl_lahir:=input.Tahun+input.Bulan+input.Tgl

	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	f, err := os.OpenFile(dir+"/error_registrasi.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	
	var toml_config config.TomlConfig
	if _, err := toml.DecodeFile(dir+"/config.toml", &toml_config); err != nil {
		fmt.Println("Error terjadi "+ err.Error())
	}

	err=nasabah.NasabahRepo.CheckNoNIK(input.NIK)
	if(err==nil){
		return 0,"No NIK Sudah terdaftar",""
	}
	err=nasabah.NasabahRepo.CheckNoTelp(input.NoHP)
	if(err==nil){
		return 0,"No Telepon Sudah Terdaftar",""
	}

	kode_kantor:=nasabah.NasabahRepo.GetKodeKantor()
	nasabah_id:=nasabah.NasabahRepo.GetMaxNasabahId()
    keyname:=toml_config.SET_ID.KEYNAME_NASABAH_ID
	templateNID:=nasabah.NasabahRepo.GetTemplateNID(keyname)
	list_tab:=nasabah.NasabahRepo.GetTabProduk()
	gen_nasabah_id,kode_k:=helpers.CreateNasabahID(kode_kantor,templateNID,nasabah_id)
	//fmt.Println(gen_nasabah_id)
	encrypt_nasabah_id:=helpers.Sha1(gen_nasabah_id)
	encrypt_password:=helpers.Sha1("123456")
	//Begin Transaction...
	tx,err_sql:= nasabah.DB.Begin()
	if(err_sql!=nil){
		fmt.Println(err_sql.Error())
	}

	_,err_sql=tx.Exec("INSERT INTO reg_pay_account VALUES (?,?,?,?,?,?)",gen_nasabah_id,input.NoHP,"0","0","0","9991")
    if(err_sql!=nil){
		tx.Rollback()
		log.Println("Gagal melakukan registrasi,terdapat kesalahan pada operasi insert reg_pay_account,desc_error : "+err_sql.Error())
	    return 0,"Kesalahan Server",err_sql.Error()
	}

	_,err_sql=tx.Exec("INSERT INTO register_pay_service VALUES (?,?,?,?,?,?,?)",
	encrypt_nasabah_id,"0017","",encrypt_password,"","4","1")
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Gagal melakukan registrasi,terdapat kesalahan pada operasi insert reg_pay_service,desc_error : "+err_sql.Error())
		return 0,"Kesalahan Server",err_sql.Error()
	}

	_,err_sql=tx.Exec("INSERT INTO nasabah (NASABAH_ID,NAMA_NASABAH,ALAMAT,TELPON,JENIS_KELAMIN,EMAIL,TEMPATLAHIR,TGLLAHIR,JENIS_ID,NO_ID,TGL_REGISTER) values(?,?,?,?,?,?,?,?,?,?,?)",
	gen_nasabah_id,input.Nama,input.Alamat,input.NoHP,input.JK,input.Email,
	input.TempatLahir,tgl_lahir,"1",input.NIK,tanggal_register)
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Gagal melakukan registrasi,terdapat kesalahan pada operasi insert nasabah,desc_error : "+err_sql.Error())
	    return 0,"Kesalahan Server",err_sql.Error()
	} 
    

	for _,v :=range list_tab {
		 kode_produk:=v.KodeProduk
		 suku_bunga:=v.SukuBungaDefault
		 //setoran_pertama:=v.SetoranPertama
		 rekening_fix:=kode_kantor+"."+kode_produk+"."+kode_k
		 _,err_sql:=tx.Exec("INSERT INTO TABUNG (NO_REKENING,NASABAH_ID,KODE_PRODUK,KODE_BI_PEMILIK,SUKU_BUNGA,TGL_REGISTER,SALDO_AKHIR,VERIFIKASI,STATUS,KODE_KANTOR,KODE_INTEGRASI,FLAG_PAY,FLAG_PAY_ECHANNEL,FLAG_IB_UPDATE) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		 rekening_fix,gen_nasabah_id,kode_produk,"876",suku_bunga,tanggal_register,"0","1","1",kode_kantor,kode_produk,"T","Y","1")

		 if(err_sql!=nil){
			 tx.Rollback()
			 log.Println("Gagal melakukan registrasi,terdapat kesalahan pada operasi insert tabung,desc_error : "+err_sql.Error())
			 return 0,"Kesalahan Server",err_sql.Error()
		 } 
	}
	 
	err_sql=tx.Commit()
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Gagal melakukan registrasi,terdapat kesalahan pada operasi commit,desc_error : "+err_sql.Error())
		return 0,"Kesalahan Server",err_sql.Error()
	}

	//end transaction...
	return 1,"Berhasil Mendaftar",""
}
