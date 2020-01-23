package usecase

import(
	"github.com/gwlkm_service/module/agen/model"
	"github.com/gwlkm_service/config"
	repo_nasabah "github.com/gwlkm_service/module/nasabah/repository"
	repo_agen "github.com/gwlkm_service/module/agen/repository"
	"github.com/gwlkm_service/helpers"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
	_"strings"
	"strconv"
	_"database/sql"
	"github.com/dustin/go-humanize"
	_"github.com/kpango/glg"
)

type AgenUseCaseImp struct{
	NasabahRepo  repo_nasabah.NasabahRepo
	AgenRepos    repo_agen.AgenRepo
	DB *sqlx.DB
}

func NewAgenUsecaseImp(NasabahRepo repo_nasabah.NasabahRepo,AgenRepos repo_agen.AgenRepo,DB *sqlx.DB)*AgenUseCaseImp{
	return &AgenUseCaseImp{
	  NasabahRepo : NasabahRepo,
	  AgenRepos   : AgenRepos,
	  DB          : DB,
	}
}

func(agen *AgenUseCaseImp) GetTabProgram()[]*model.TabProduk {
	list_produk:=agen.AgenRepos.GetTabProgram()
	return list_produk
}

func(agen *AgenUseCaseImp) GetListAnggota(nasabah_id string)[]*model.ListAnggota{
	list_anggota:=agen.AgenRepos.GetListAnggota(nasabah_id)
	no_rekening:=agen.AgenRepos.GetNoRekeningByJenisTabungan(nasabah_id,"SPK")
	fmt.Println(no_rekening)
	saldo:=agen.AgenRepos.GetSetoranPertamaJenisTabungan("SPK")
	fmt.Println(saldo)
	return list_anggota
}


//mengandung operasi sql besar.....
func(agen *AgenUseCaseImp) CreateTabProgram(inputTab model.InputTab)(string,string,string){
	deskripsi,suku_bunga,biaya_tab:=agen.AgenRepos.GetBiayaTabProgram(inputTab.KodeProduk)
	fmt.Println("Biaya Tabungan")
	fmt.Println(biaya_tab)
	produk:=agen.NasabahRepo.GetOneTabProduk()
	rekening_agen,saldo_akhir,err:=agen.NasabahRepo.GetNasabahTabPayment(inputTab.IDUser,produk.KodeProduk)
	if(err!=nil){
		return "0","Anda tidak punya rekening sukarela",err.Error()
	}
	fmt.Println(saldo_akhir.Float64)
	///
	if(biaya_tab>saldo_akhir.Float64){
		return "0","Anda harus punya saldo minimal Rp."+humanize.Comma(int64(biaya_tab))+" untuk membuka tabungan "+deskripsi,""
	}

	err=agen.AgenRepos.CheckNasabahID(inputTab.IDNasabah)
	if(err!=nil){
		return "0","Nasabah ID "+inputTab.IDNasabah+" tidak memiliki tabungan",err.Error()
	}

	jumlah_tabungan:=agen.AgenRepos.GetCountTabungan(inputTab.IDNasabah)
	jumlah_tabungan=jumlah_tabungan+1

	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	f, err := os.OpenFile(dir+"/error_create_tab_program.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	
	var toml_config config.TomlConfig
	if _, err := toml.DecodeFile(dir+"/config.toml", &toml_config); err != nil {
		fmt.Println("Error terjadi "+ err.Error())
	}

	keyname:=toml_config.SET_ID.KEYNAME_NOREK
	templateRekening:=agen.NasabahRepo.GetTemplateNID(keyname)
	kode_kantor:=agen.NasabahRepo.GetKodeKantor()
	rekening_baru:=helpers.CreateTabProgram(inputTab.IDNasabah,inputTab.KodeProduk,kode_kantor,strconv.Itoa(jumlah_tabungan),templateRekening)

	currentTime := time.Now()
	tanggal_register:=currentTime.Format("2006-01-02")
	timestamp := currentTime.Format("2006-01-02 15:04:05")
	time := currentTime.Format("15:04:05")
	kwitansi := currentTime.Format("20060102150405")  


	///mulai transaksi
	tx,err_sql:= agen.DB.Begin()

	
  ///proses insert tabungan baru
	_,err_sql=tx.Exec("INSERT INTO TABUNG (NO_REKENING,NASABAH_ID,KODE_PRODUK,KODE_BI_PEMILIK,SUKU_BUNGA,TGL_REGISTER,SALDO_AKHIR,VERIFIKASI,STATUS,KODE_KANTOR,KODE_INTEGRASI,FLAG_PAY,FLAG_PAY_ECHANNEL,FLAG_IB_UPDATE) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	rekening_baru,inputTab.IDNasabah,inputTab.KodeProduk,"874",suku_bunga,tanggal_register,biaya_tab,"1","1",kode_kantor,inputTab.KodeProduk,"T","Y","1")
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert tabung , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	
	desc_tabung:="Registrasi tabungan "+deskripsi+" Rek "+rekening_baru+" oleh "+rekening_agen

	///insert tabtrans pertama
	trans_pertama:=agen.NasabahRepo.GetMaxTransID()
	trans_pertama=helpers.CreateTransID(trans_pertama)

    if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	  "pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans)"+
	  " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_pertama,tanggal_register,rekening_agen,"204","200",biaya_tab,
	   "0",kwitansi,"9999",desc_tabung,"1","T",rekening_baru,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		  tx.Rollback()
		  log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert tabtrans untuk agen , desc :"+err_sql.Error())
		  return "0","Kesalahan Server",err_sql.Error()
	}

   ///insert transaksi master pertama
	trans_id_first_int,_:=strconv.Atoi(trans_pertama)
	trans_id_first_master:=trans_id_first_int+1

	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	trans_id_first_master,"TAB",kwitansi,tanggal_register,desc_tabung,"VIP",trans_id_first_master,
	"9999","",kode_kantor,"5",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert tabmaster untuk agen , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	 ///insert transaksi detail tahap 1 split pertama..
	 transaksi_detail_first:=trans_id_first_master+1
	 if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",transaksi_detail_first,trans_id_first_master,"20101",biaya_tab,"0","",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert transaksi_detail debit untuk agen , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
	  }


	 ///insert transaksi detail tahap 2 split pertama..
	 transaksi_detail_first_x:=transaksi_detail_first+1
	 if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",transaksi_detail_first_x,trans_id_first_master,"20101","0",biaya_tab,"",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert transaksi_detail kredit untuk agen , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
	  }


   //insert tabtrans kedua ..
   trans_id:=transaksi_detail_first_x+1

	 if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	 "pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans)"+
	 " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id,tanggal_register,rekening_baru,"204","100",biaya_tab,
	  "0",kwitansi,"9999",desc_tabung,"1","T",rekening_agen,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		 tx.Rollback()
		 log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert tabtrans  untuk user , desc :"+err_sql.Error())
		 return "0","Kesalahan Server",err_sql.Error()
	 }

	 trans_id_2:=trans_id+1

	  ///insert transaksi master  kedua..
	  if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	  trans_id_2,"TAB",kwitansi,tanggal_register,desc_tabung,"VIP",trans_id_2,
	  "9999","",kode_kantor,"5",""); err_sql != nil {
		  tx.Rollback()
		  log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert transaksi_master  untuk user , desc :"+err_sql.Error())
		  return "0","Kesalahan Server",err_sql.Error()
	  }

	  ///insert transaksi detail tahap 1 split keduaa..
	  transaksi_detail_1:=trans_id_2+1
	  if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",transaksi_detail_1,trans_id_2,"20101",biaya_tab,"0","",kode_kantor); err_sql != nil {
			 tx.Rollback()
			 log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert detail transaksi debit  untuk user , desc :"+err_sql.Error())
			 return "0","Kesalahan Server",err_sql.Error()
	   }
		
	   ///insert transaksi detail tahap 2 split kedua..
      transaksi_detail_2:=transaksi_detail_1+1
	  if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	     transaksi_detail_2,trans_id_2,"20101","0",biaya_tab,"",kode_kantor); err_sql != nil {
		 tx.Rollback()
		 log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert detail transaksi kredit  untuk user , desc :"+err_sql.Error())
		 return "0","Kesalahan Server",err_sql.Error()
	 } 

     ///update saldo rekening baru
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = ? WHERE NO_REKENING =? ",biaya_tab,rekening_baru); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat update saldo untuk user , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

   //update saldo rekening agem
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE NO_REKENING =? ",biaya_tab,rekening_agen); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat update saldo untuk agen , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	keyname_tab:="FEE_TAB_PROGRAM_AGEN_REK_"+inputTab.KodeProduk
	fmt.Println(keyname_tab)
	fmt.Println(inputTab.KodeProduk)
	biaya_admin_tabungan:=agen.AgenRepos.GetBiayaAdminTabungan(keyname_tab,inputTab.KodeProduk)
	fmt.Println("Biaya Admin Tabungan")
	fmt.Println(biaya_admin_tabungan)
	desc_admin:="Fee Registrasi Tabungan "+deskripsi+" Rek "+rekening_baru+" oleh "+rekening_agen

	///insert tabtrans biaya admin tabungan
	trans_id_admin_tabungan:=transaksi_detail_2+1

	if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,kode_kantor,jam,waktu,tgl_real_trans)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id_admin_tabungan,tanggal_register,rekening_agen,"204","100",biaya_admin_tabungan,
	 "0",kwitansi,"9999",desc_admin,"1","T",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert tabtrans untuk agen(fee_admin) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

    ///insert transaksi master biaya admin tabungan
	trans_id_tabungan_master:=trans_id_admin_tabungan+1
	
	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	trans_id_tabungan_master,"TAB",kwitansi,tanggal_register,desc_admin,"VIP",trans_id_tabungan_master,
	"9999","",kode_kantor,"5",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert transaksi_master untuk agen(fee_admin) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	//insert transaksi detail debit biaya admin tabungan...
	trans_id_detail_debit_admin_tabungan:=trans_id_tabungan_master+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",trans_id_detail_debit_admin_tabungan,trans_id_tabungan_master,"20101",biaya_admin_tabungan,"0","",kode_kantor); err_sql != nil {
		   tx.Rollback()
		   log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert transaksi_detail debit untuk agen(fee_admin) , desc :"+err_sql.Error())
		   return "0","Kesalahan Server",err_sql.Error()
	 }


	 //insert transaksi detail kredit biaya admin tabungan...
	trans_id_detail_kredit_admin_tabungan:=trans_id_detail_debit_admin_tabungan+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",trans_id_detail_kredit_admin_tabungan,trans_id_tabungan_master,"20101","0",biaya_admin_tabungan,"",kode_kantor); err_sql != nil {
		   tx.Rollback()
		   log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat insert transaksi_detail kredit untuk agen(fee_admin) , desc :"+err_sql.Error())
		   return "0","Kesalahan Server",err_sql.Error()
	 }

    //update rekening agen biaya admin tabungan...
	 if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR  + ? WHERE NO_REKENING =? ",biaya_admin_tabungan,rekening_agen); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat update saldo untuk agen(fee_admin) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	
	///update saldo perkiraan biaya admin tabungan...
	if _, err_sql = tx.Exec("UPDATE PERKIRAAN SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE KODE_PERK =? ",biaya_admin_tabungan,"20101"); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada saat update saldo perkiraan untuk agen(fee_admin) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///lakukan commit
	err_sql=tx.Commit()
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+inputTab.IDUser+" gagal membuat tabungan baru,Terdapat kesalahan pada commit , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

   return "1","Berhasil membuka tabungan "+deskripsi,""
}




///warning ....mengandung operasi sql besar...
func(agen *AgenUseCaseImp)RegisterViaAgen(input model.InputRegisterViaAgen)(string,string,string){
	var rekening_pokok string
	produk:=agen.NasabahRepo.GetOneTabProduk()
	rekening_agen,saldo_akhir,err:=agen.NasabahRepo.GetNasabahTabPayment(input.NasabahID,produk.KodeProduk)
	if(err!=nil){
		return "0","Anda tidak mempunyai rekening sukarela",err.Error()
	}
	saldo_akhir_pendaftar:=saldo_akhir.Float64
	biaya_admin:=agen.AgenRepos.GetAgenKeyValue("TAB_ECH_KOMUNITAS_ADM_ANGGOTA_DEFAULT")
	biaya_tabungan:=agen.AgenRepos.GetSumTabungan()
	total_biaya_pendaftaran:=biaya_admin+biaya_tabungan
	saldo_split_a:=saldo_akhir_pendaftar-total_biaya_pendaftaran
	if(saldo_akhir_pendaftar<total_biaya_pendaftaran){
		return "0","Saldo Anda tidak mencukupi,minimal saldo harus Rp."+humanize.Comma(int64(total_biaya_pendaftaran)),""
	}



	
	currentTime := time.Now()
	tanggal_register:=currentTime.Format("2006-01-02")
	timestamp := currentTime.Format("2006-01-02 15:04:05")  
	kwitansi := currentTime.Format("20060102150405")  
	time := currentTime.Format("15:04:05")    
	tgl_lahir:=input.Tahun+input.Bulan+input.Tgl

	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	f, err := os.OpenFile(dir+"/error_registrasi_via_agen.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	
	var toml_config config.TomlConfig
	if _, err := toml.DecodeFile(dir+"/config.toml", &toml_config); err != nil {
		fmt.Println("Error terjadi "+ err.Error())
	}

	err=agen.NasabahRepo.CheckNoNIK(input.NIK)
	if(err==nil){
		return "0","No NIK Sudah terdaftar",""
	}
	err=agen.NasabahRepo.CheckNoTelp(input.NoHP)
	if(err==nil){
		return "0","No Telepon Sudah Terdaftar",""
	}

	kode_kantor:=agen.NasabahRepo.GetKodeKantor()
	nasabah_id:=agen.NasabahRepo.GetMaxNasabahId()
    keyname:=toml_config.SET_ID.KEYNAME_NASABAH_ID
	templateNID:=agen.NasabahRepo.GetTemplateNID(keyname)
	list_tab:=agen.NasabahRepo.GetTabProduk()
	gen_nasabah_id,kode_k:=helpers.CreateNasabahID(kode_kantor,templateNID,nasabah_id)
	encrypt_nasabah_id:=helpers.Sha1(gen_nasabah_id)
	encrypt_password:=helpers.Sha1("123456")

	//mulai transaksi
	tx,err_sql:= agen.DB.Begin()

	//proses pembuatan reg_pay_account..
	_,err_sql=tx.Exec("INSERT INTO reg_pay_account VALUES (?,?,?,?,?,?)",gen_nasabah_id,input.NoHP,"0","0","0","9991")
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert reg_pay_account , desc :"+err_sql.Error())
	    return "0","Kesalahan Server",err_sql.Error()
	}

   ///pembuatan reg_pay_service...
	_,err_sql=tx.Exec("INSERT INTO register_pay_service VALUES (?,?,?,?,?,?,?)",
	encrypt_nasabah_id,"0017","",encrypt_password,"","4","1")
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert reg_pay_service , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

   ///proses insert data nasabah..
	_,err_sql=tx.Exec("INSERT INTO NASABAH (NASABAH_ID,NAMA_NASABAH,ALAMAT,TELPON,JENIS_KELAMIN,EMAIL,TEMPATLAHIR,TGLLAHIR,JENIS_ID,NO_ID,TGL_REGISTER,REFERAL_ID) values(?,?,?,?,?,?,?,?,?,?,?,?)",
	gen_nasabah_id,input.Nama,input.Alamat,input.NoHP,input.JK,input.Email,
	input.TempatLahir,tgl_lahir,"1",input.NIK,tanggal_register,input.NasabahID)
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert nasabah , desc :"+err_sql.Error())
	    return "0","Kesalahan Server",err_sql.Error()
	} 
    i:=0;
	///pembuatan tabungan berdasarkan tabproduk
	no_rekening_simpok:=""
	no_rekening_simwa:=""

	for _,v :=range list_tab {
		rekening_fix:=""
		kode_produk:=v.KodeProduk
		suku_bunga:=v.SukuBungaDefault
		if(i==0){
		   rekening_fix=kode_kantor+"."+kode_produk+"."+kode_k
		   rekening_pokok=rekening_fix
		} else {
		   rekening_fix=kode_kantor+"."+kode_produk+"."+kode_k
		}

		if(v.Jenis=="SPK"){
			no_rekening_simpok=kode_kantor+"."+kode_produk+"."+kode_k
		}

		if(v.Jenis=="SWJ"){
			no_rekening_simwa=kode_kantor+"."+kode_produk+"."+kode_k
		}

		_,err_sql:=tx.Exec("INSERT INTO TABUNG (NO_REKENING,NASABAH_ID,KODE_PRODUK,KODE_BI_PEMILIK,SUKU_BUNGA,TGL_REGISTER,SALDO_AKHIR,VERIFIKASI,STATUS,KODE_KANTOR,KODE_INTEGRASI,FLAG_PAY,FLAG_PAY_ECHANNEL,FLAG_IB_UPDATE) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		rekening_fix,gen_nasabah_id,kode_produk,"876",suku_bunga,tanggal_register,"0","1","1",kode_kantor,kode_produk,"T","Y","1")
		//rekening_pokok=rekening_fix
		i=i+1
		if(err_sql!=nil){
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert tabung , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		} 
   }

   ///awal proses split pertama

   trans_id:=agen.NasabahRepo.GetMaxTransID()
   trans_id=helpers.CreateTransID(trans_id)
   
   desc_trans:="Pengambilan Rekening simpanan dari rekening "+rekening_agen+" ke "+rekening_pokok

   //insert tabtrans pertama split pertama....
   if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id,tanggal_register,rekening_agen,"204","200",total_biaya_pendaftaran,
	 "0",kwitansi,"9999",desc_trans,"1","T",rekening_pokok,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert tabtrans untuk agen , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}
	
   trans_id_int,_:=strconv.Atoi(trans_id)
   trans_id_2:=trans_id_int+1

  ///insert tabtrans kedua split pertama...
   if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
   "pokok,adm,kuitansi,userid,keterangan,verifikasi,modul_id_source,trans_id_source,tob,kode_kantor,jam,waktu,tgl_real_trans)"+
   " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id_2,tanggal_register,rekening_pokok,"104","100",total_biaya_pendaftaran,
	"0",kwitansi,"9999",desc_trans,"1","TAB",trans_id,"O",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
	   tx.Rollback()
	   log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert tabtrans untuk user , desc :"+err_sql.Error())
	   return "0","Kesalahan Server",err_sql.Error()
   }

  transaksi_master:=trans_id_2+1

   ///insert transaksi master split pertama...
   if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
   transaksi_master,"TAB",kwitansi,tanggal_register,desc_trans,"VIP",transaksi_master,
   "9999","",kode_kantor,"5",""); err_sql != nil {
	   tx.Rollback()
	   log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_master untuk agen , desc :"+err_sql.Error())
	   return "0","Kesalahan Server",err_sql.Error()
   }


  ///insert transaksi detail tahap 1 split pertama..
   transaksi_detail_1:=transaksi_master+1
   
   if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",transaksi_detail_1,transaksi_master,"20101",total_biaya_pendaftaran,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_detail debit untuk agen , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


   ///insert transaksi detail tahap 2 split pertama..
	transaksi_detail_2:=transaksi_detail_1+1

	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	 transaksi_detail_2,transaksi_master,"20101","0",total_biaya_pendaftaran,"",kode_kantor); err_sql != nil {
		 tx.Rollback()
		 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_detail kredit untuk agen , desc :"+err_sql.Error())
		 return "0","Kesalahan Server",err_sql.Error()
	 }

	///update saldo tahap 1 split pertama
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = ? WHERE NO_REKENING =? ",saldo_split_a,rekening_agen); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo untuk agen , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///update saldo tahap 2 split pertama
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = ? WHERE NO_REKENING =? ",total_biaya_pendaftaran,rekening_pokok); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo untuk user , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	//akhhir proses split pertama
	biaya_admin_agt_ref_default:=agen.AgenRepos.GetAgenKeyValue("TAB_ECH_KOMUNITAS_ADM_ANGGOTA_REFERRAL_DEFAULT")
	

	///awal proses split kedua
	tabtrans_2:=transaksi_detail_2+1
	desc_admin:="Adm Setoran awal registerasi "+rekening_pokok

	//insert tabtrans split kedua
	if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,sandi_trans,kode_perk_ob,kode_kantor,jam,waktu,tgl_real_trans)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_2,tanggal_register,rekening_pokok,"206","200",biaya_admin_agt_ref_default,
	 "0",kwitansi,"9999",desc_admin,"1","O",rekening_pokok,"08","4010202",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert tabtrans untuk user(split kedua) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	transaksi_master_2:=tabtrans_2+1


  ///insert transaksi master split kedua
   if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
   transaksi_master_2,"TAB",kwitansi,tanggal_register,desc_admin,"TAB",tabtrans_2,
   "9999","",kode_kantor,"0",""); err_sql != nil {
	   tx.Rollback()
	   log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_master untuk user(split kedua) , desc :"+err_sql.Error())
	   return "0","Kesalahan Server",err_sql.Error()
   }

   ///insert transaksi detail tahap 1 split kedua
   transaksi_detail_3:=transaksi_master_2+1
   if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	transaksi_detail_3,transaksi_master_2,"20101",biaya_admin_agt_ref_default,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_detail debit untuk user(split kedua) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///insert transaksi detail tahap 2 split kedua
	transaksi_detail_4:=transaksi_detail_3+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	transaksi_detail_4,transaksi_master_2,"4010202","0",biaya_admin_agt_ref_default,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_detail kredit untuk user(split kedua) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///
	saldo_akhir_perkiraan:=agen.AgenRepos.GetSaldoAkhirPerkiraan("20101")
	update_saldo_akhir_perkiraan:=saldo_akhir_perkiraan+biaya_admin_agt_ref_default
	saldo_split_b:=total_biaya_pendaftaran-biaya_admin_agt_ref_default

	///update saldo tahap 1 split kedua
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = ? WHERE NO_REKENING =? ",saldo_split_b,rekening_pokok); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo untuk user(split kedua) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	///update saldo tahap 2 split kedua
	if _, err_sql = tx.Exec("UPDATE PERKIRAAN SET SALDO_AKHIR = ? WHERE KODE_PERK =? ",update_saldo_akhir_perkiraan,"20101"); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo perkiraan untuk user(split kedua) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	//akhir proses split kedua


   //mulai proses split ketiga
	biaya_komunitas_default:=agen.AgenRepos.GetAgenKeyValue("TAB_ECH_KOMUNITAS_ADM_ANGGOTA_AGEN_DEFAULT")
	desc_fee:="Fee Agen Registrasi : "+rekening_pokok
	trans_id_3:=transaksi_detail_4+1

	//insert tabtrans split ketiga...
	if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,sandi_trans,kode_kantor,jam,waktu,tgl_real_trans)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id_3,tanggal_register,rekening_pokok,"204","200",biaya_komunitas_default,
    "0",kwitansi,"9999",desc_fee,"1","O",rekening_agen,"08",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert tabtrans untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
	}

	trans_terakhir:=trans_id_3+1
	desc_ob:="Setoran Tabungan OB dari Rekening Tabungan "+rekening_pokok
	 ///insert tabtrans kedua split pertama...
	 if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	 "pokok,adm,kuitansi,userid,keterangan,verifikasi,sandi_trans,modul_id_source,trans_id_source,tob,kode_kantor,jam,waktu,tgl_real_trans)"+
	 " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_terakhir,tanggal_register,rekening_agen,"104","100",biaya_komunitas_default,
	  "0",kwitansi,"9999",desc_ob,"1","08","TAB",trans_id_3,"O",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		 tx.Rollback()
		 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert tabtrans terakhir , desc :"+err_sql.Error())
		 return "0","Kesalahan Server",err_sql.Error()
	 }

	//insert transaksi master tahap split ketiga
	transaksi_master_3:=trans_terakhir+1
	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	transaksi_master_3,"TAB",kwitansi,tanggal_register,desc_fee,"TAB",trans_id_3,
	"9999","",kode_kantor,"0",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_master untuk user(split ketiga) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	//insert transaksi detail 1  tahap split ketiga
	transaksi_detail_5:=transaksi_detail_4+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	transaksi_detail_5,transaksi_master_3,"20101",biaya_komunitas_default,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_detail debit untuk user(split ketiga) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	//insert transaksi detail   tahap split ketiga
	transaksi_detail_6:=transaksi_detail_5+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	transaksi_detail_6,transaksi_master_3,"4010202","0",biaya_komunitas_default,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert transaksi_detail kredit untuk user(split ketiga) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	
  
	

	saldo_split_c:=saldo_split_a+biaya_komunitas_default
	saldo_split_d:=saldo_split_b-biaya_komunitas_default

	///update saldo tahap 1 split ketiga
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = ? WHERE NO_REKENING =? ",saldo_split_c,rekening_agen); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update sado untuk agen(split ketiga), desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	///update saldo tahap 2 split ketiga
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = ? WHERE NO_REKENING =? ",saldo_split_d,rekening_pokok); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo debit untuk user(split ketiga) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}
	///akhir split ketiga

	

	//update poin agen...
	poin:=toml_config.Point_trans.POINT_REG_AGT_BYAGENT
    if _, err_sql = tx.Exec("UPDATE nasabah SET total_point = total_point + ? WHERE NASABAH_ID = ? ",poin,input.NasabahID); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update point untuk agen(split ketiga) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///perubahan
	setoran_pertama_simpanan_wajib:=agen.AgenRepos.GetSetoranPertamaJenisTabungan("SWJ")
	setoran_pertama_simpanan_pokok:=agen.AgenRepos.GetSetoranPertamaJenisTabungan("SPK")
	//total_setoran_simpanan:=setoran_pertama_simpanan_pokok+setoran_pertama_simpanan_wajib

	//no_rekening_simwa:=agen.AgenRepos.GetNoRekeningByJenisTabungan(gen_nasabah_id,"SWJ")
	//fmt.Println("no_"+no_rekening_simwa)
	//no_rekening_simpok:=agen.AgenRepos.GetNoRekeningByJenisTabungan(gen_nasabah_id,"SPK")
	//fmt.Println("no_"+no_rekening_simpok)


	///_,saldo_akhir_anggota,_:=agen.NasabahRepo.GetNasabahTabPayment(gen_nasabah_id,produk.KodeProduk)

// 	if(saldo_akhir_anggota.Float64>total_setoran_simpanan){

		desc_simpanan_wajib:="Distribusi ke Simpanan wajib dari Rek "+rekening_pokok
		tabtrans_wjb:=transaksi_detail_6+1
		if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
			"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,sandi_trans,kode_kantor,jam,waktu,tgl_real_trans)"+
			" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_wjb,tanggal_register,rekening_pokok,"204","200",setoran_pertama_simpanan_wajib,
			"0",kwitansi,"9999",desc_simpanan_wajib,"1","O",no_rekening_simwa,"08",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert tabtrans untuk user , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		 }
		 
		 
		 tabtrans_wjb_1:=tabtrans_wjb+1
		 if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
			"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,sandi_trans,kode_kantor,jam,waktu,tgl_real_trans)"+
			" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_wjb_1,tanggal_register,no_rekening_simwa,"204","100",setoran_pertama_simpanan_wajib,
			"0",kwitansi,"9999",desc_simpanan_wajib,"1","O",rekening_pokok,"08",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert tabtrans untuk user , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		 }

		 transaksi_master_wjb:=tabtrans_wjb_1+1
		if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
		transaksi_master_wjb,"TAB",kwitansi,tanggal_register,desc_simpanan_wajib,"TAB",tabtrans_wjb_1,
		"9999","",kode_kantor,"0",""); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert distribusi simpanan wajib transaksi_master untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}


		transaksi_detail_wjb_1:=transaksi_master_wjb+1
		if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		transaksi_detail_wjb_1,transaksi_master_wjb,"20101",setoran_pertama_simpanan_wajib,"0","",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert transaksi_detail debit untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}


		//insert transaksi detail   tahap split ketiga
		transaksi_detail_wjb_2:=transaksi_detail_wjb_1+1
		if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		transaksi_detail_wjb_2,transaksi_master_wjb,"30102","0",setoran_pertama_simpanan_wajib,"",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert transaksi_detail kredit untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}

		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE NO_REKENING =? ",setoran_pertama_simpanan_wajib,rekening_pokok); 
	    err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update sado untuk agen(split ketiga), desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
	    }


		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE NO_REKENING =? ",setoran_pertama_simpanan_wajib,no_rekening_simwa); 
			err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo debit untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}


		desc_simpanan_pokok:="Distribusi ke Simpanan Pokok dari Rek "+rekening_pokok

		tabtrans_pokok:=transaksi_detail_wjb_2+1
		if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
			"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,sandi_trans,kode_kantor,jam,waktu,tgl_real_trans)"+
			" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_pokok,tanggal_register,rekening_pokok,"204","200",setoran_pertama_simpanan_pokok,
			"0",kwitansi,"9999",desc_simpanan_pokok,"1","O",no_rekening_simpok,"08",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert tabtrans untuk user , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		 }


		 tabtrans_pokok_1:=tabtrans_pokok+1
		 if _, err_sql = tx.Exec("insert into tabtrans(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
			"pokok,adm,kuitansi,userid,keterangan,verifikasi,tob,no_rekening_vs,sandi_trans,kode_kantor,jam,waktu,tgl_real_trans)"+
			" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_pokok_1,tanggal_register,no_rekening_simpok,"204","100",setoran_pertama_simpanan_pokok,
			"0",kwitansi,"9999",desc_simpanan_pokok,"1","O",rekening_pokok,"08",kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert tabtrans untuk user , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		 }


		transaksi_master_pokok:=tabtrans_pokok_1+1
		if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
		transaksi_master_pokok,"TAB",kwitansi,tanggal_register,desc_simpanan_pokok,"TAB",tabtrans_pokok_1,
		"9999","",kode_kantor,"0",""); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat insert distribusi simpanan wajib transaksi_master untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}


		transaksi_detail_pokok_1:=transaksi_master_pokok+1
		if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		transaksi_detail_pokok_1,transaksi_master_pokok,"20101",setoran_pertama_simpanan_pokok,"0","",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert transaksi_detail debit untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}


		//insert transaksi detail   tahap split ketiga
		transaksi_detail_pokok_2:=transaksi_detail_pokok_1+1
		if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		transaksi_detail_pokok_2,transaksi_master_pokok,"30101","0",setoran_pertama_simpanan_pokok,"",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat distribusi simpanan wajib insert transaksi_detail kredit untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}

		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE NO_REKENING =? ",setoran_pertama_simpanan_pokok,rekening_pokok); 
	    err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update sado untuk agen(split ketiga), desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
	    }


		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE NO_REKENING =? ",setoran_pertama_simpanan_pokok,no_rekening_simpok); 
			err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat update saldo debit untuk user(split ketiga) , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}
		 
//	}

	//perubahan

	//commit sql, akhir dari proses regitrasi by agen...
	err_sql=tx.Commit()
	if(err_sql!=nil){
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan register via agen ,Terdapat kesalahan pada saat commit, desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

    return "1","Pendaftaran Berhasil",""
}






