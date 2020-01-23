package usecase

import (
	"github.com/gwlkm_service/module/payment/model"
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
	_"github.com/dustin/go-humanize"
	_"github.com/kpango/glg"
)

type PaymentUseCaseImp struct{
	NasabahRepo  repo_nasabah.NasabahRepo
	AgenRepos    repo_agen.AgenRepo
	DB *sqlx.DB
}

func NewPaymentUsecaseImp(NasabahRepo repo_nasabah.NasabahRepo,AgenRepos repo_agen.AgenRepo,DB *sqlx.DB)*PaymentUseCaseImp{
	return &PaymentUseCaseImp{
	  NasabahRepo : NasabahRepo,
	  AgenRepos   : AgenRepos,
	  DB          : DB,
	}
}

func(payment *PaymentUseCaseImp)PayHandler(input model.InputPay)(string,string,string){
	currentTime := time.Now()

	tanggal_register:=currentTime.Format("2006-01-02")
	timestamp := currentTime.Format("2006-01-02 15:04:05")  
	time := currentTime.Format("15:04:05")
	kode_kantor:=payment.NasabahRepo.GetKodeKantor()

	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	fmt.Println(input.Adm)

	f, err := os.OpenFile(dir+"/error_payment.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	
	var toml_config config.TomlConfig

	if _, err := toml.DecodeFile(dir+"/config.toml", &toml_config); err != nil {
		fmt.Println("Gagal membaca file konfigurasi "+ err.Error())
	}

	produk:=payment.NasabahRepo.GetOneTabProduk()
	rekening_nasabah,saldo_akhir,err:=payment.NasabahRepo.GetNasabahTabPayment(input.NasabahID,produk.KodeProduk)

	if(err!=nil){
		return "0","Anda tidak mempunyai rekening sukarela",err.Error()
	}

	saldo_akhir_nasabah:=saldo_akhir.Float64
	nominal_produk:=float64(input.Nominal)

	if(nominal_produk>saldo_akhir_nasabah){
		return "0","Saldo anda tidak mencukupi",""
	}

	trans_id:=payment.NasabahRepo.GetMaxTransID()
	///mapping:=payment.NasabahRepo.GetPayProductMapping("400000")
	desc_transaksi:=input.Keterangan+" ke IDPEL "+input.Tujuan

	//mulai transaksi
	trans_id=helpers.CreateTransID(trans_id)
	tx,err_sql:= payment.DB.Begin()

    ///memasukkan tabtrans pertama...
	if _, err_sql = tx.Exec("insert into TABTRANS(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
	"pokok,adm,kuitansi,userid,keterangan,kode_kantor,jam,waktu,tgl_real_trans)"+
	" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)",trans_id,tanggal_register,rekening_nasabah,"301","200",nominal_produk,
	 "0",input.Reffid,"9990",desc_transaksi,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada insert tabtrans untuk user , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	trans_id_int,_:=strconv.Atoi(trans_id)
	trans_id_2:=trans_id_int+1

   ///memasukkan transaksi_master...
	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	trans_id_2,"TAB",input.Reffid,tanggal_register,desc_transaksi,"TAB","139128003",
	"9990","",kode_kantor,"0",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada insert transaksi_master untuk user , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	///
	trans_detail_1:=trans_id_2+1
	trans_detail_2:=trans_detail_1+1
	
	perkiraan_sukarela:=toml_config.Perkiraan.SUKARELA_LKM
	perkiraan_deposit:=toml_config.Perkiraan.DEPOSIT_ECHANNEL
	perkiraan_pendapatan:=toml_config.Perkiraan.PENDAPATAN_NON_OPERASIONAL

	///insert transaksi detail pertama
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	trans_detail_1,trans_id_2,perkiraan_sukarela,nominal_produk,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada insert transaksi detail debit untuk user , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

   fmt.Println(trans_detail_2)
	///insert transaksi detail kedua
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	trans_detail_2,trans_id_2,perkiraan_deposit,"0",nominal_produk,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada insert transaksi detail kredit untuk user , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}
	fmt.Println("Rekening Nasabah Anda :"+rekening_nasabah)
    ///update saldo  user
		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE NO_REKENING =? ",nominal_produk,rekening_nasabah); 
		err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada update saldo untuk user , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
		}


    ///update saldo perkiraan sukarela
	if _, err_sql = tx.Exec("UPDATE perkiraan SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE KODE_PERK =? ",nominal_produk,perkiraan_sukarela); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada update saldo perkiraan sukarela , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	///update saldo perkiraan deposit
	if _, err_sql = tx.Exec("UPDATE perkiraan SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE KODE_PERK =? ",nominal_produk,perkiraan_deposit); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada update saldo perkiraan deposit , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	referal_id,err:=payment.AgenRepos.GetReferalID(input.NasabahID)

	adm:=float64(input.Adm)
	fee_user:=(adm/100)*float64(toml_config.Fee_pay_pulsa.FEE_USER)
	fee_agen_user:=(adm/100)*float64(toml_config.Fee_pay_pulsa.FEE_AGEN_USER)
	fee_lkm:=(adm/100)*float64(toml_config.Fee_pay_pulsa.FEE_LKM)+fee_agen_user
	desc_user:="Fee user "+input.Keterangan+" ke IDPEL "+input.Tujuan+" dari rekening "+rekening_nasabah
	kwitansi2:="TINTCR"+input.Reffid
	fmt.Println(fee_user)
	fmt.Println(fee_lkm)
	fmt.Println(fee_agen_user)
	

	if(referal_id!=""){

		fee_lkm=(adm/100)*float64(toml_config.Fee_pay_pulsa.FEE_LKM)
		produk_2:=payment.NasabahRepo.GetOneTabProduk()
		rekening_agen,_,err:=payment.NasabahRepo.GetNasabahTabPayment(referal_id,produk_2.KodeProduk)

		if(err!=nil){
			tx.Rollback()
			log.Println("Agen dengan nasabah_id "+rekening_agen+" ini tidak punya rekening sukarela,desc error "+err.Error())
		}

		kwitansi:="TINTCR"+input.Reffid
		desc_agen:="Fee agen  "+input.Keterangan+" ke IDPEL "+input.Tujuan+" dari rekening "+rekening_nasabah
		desc_agen_client:="Fee agen  "+input.Keterangan+" ke IDPEL "+input.Tujuan

		fmt.Println(desc_agen)

		tabtrans_2_x:=trans_detail_2+1

		if _, err_sql = tx.Exec("insert into TABTRANS(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
			 "pokok,adm,kuitansi,userid,keterangan,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans)"+
			 " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_2_x,tanggal_register,rekening_nasabah,"204","100",fee_agen_user,
			 "0",kwitansi,"9999",desc_agen_client,rekening_agen,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			 tx.Rollback()
			 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert tabtrans untuk agen , desc :"+err_sql.Error())
			 return "0","Kesalahan Server",err_sql.Error()
		}
		
		///insert tabtrans untuk fee agen
       tabtrans_2:=tabtrans_2_x+1
	   if _, err_sql = tx.Exec("insert into TABTRANS(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
			"pokok,adm,kuitansi,userid,keterangan,trans_id_source,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans)"+
			" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_2,tanggal_register,rekening_agen,"204","100",fee_agen_user,
			"0",kwitansi,"9999",desc_agen,tabtrans_2_x,rekening_nasabah,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert tabtrans untuk agen , desc :"+err_sql.Error())
			return "0","Kesalahan Server",err_sql.Error()
	   }
	   ///insert tab_master untuk fee agen
	   transaksi_master_agen:=tabtrans_2+1
	   if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	   transaksi_master_agen,"TAB",kwitansi,tanggal_register,desc_agen,"TAB","139198114",
	   "9999","",kode_kantor,"0",""); err_sql != nil {
			 tx.Rollback()
			 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi_master untuk agen , desc :"+err_sql.Error())
		   return "0","Kesalahan Server",err_sql.Error()
	   }

	   ///insert transaksi detail 1 untuk fee agen
		 transaksi_detail_1_agen:=transaksi_master_agen+1
		 fmt.Println(transaksi_detail_1_agen)
	   if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
	      transaksi_detail_1_agen,transaksi_master_agen,perkiraan_sukarela,fee_agen_user,"0","",kode_kantor); err_sql != nil {
		 tx.Rollback()
		 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi_detail debit untuk agen , desc :"+err_sql.Error())
		 return "0","Kesalahan Server",err_sql.Error()
		}
		
		///insert transaksi detail 2 untuk fee agen
		trans_detail_2=transaksi_detail_1_agen+1

		if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		   trans_detail_2,transaksi_master_agen,perkiraan_sukarela,"0",fee_agen_user,"",kode_kantor); err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi_detail kredit untuk agen , desc :"+err_sql.Error())
		  return "0","Kesalahan Server",err_sql.Error()
		}

		///update saldo agen
		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE NO_REKENING =? ",fee_agen_user,rekening_agen); 
	      err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat update saldo agen, desc :"+err_sql.Error())
		  return "0","Kesalahan Server",err_sql.Error()
		}

		if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE NO_REKENING =? ",fee_agen_user,rekening_nasabah); 
	      err_sql != nil {
			tx.Rollback()
			log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat update saldo agen, desc :"+err_sql.Error())
		  return "0","Kesalahan Server",err_sql.Error()
		}
		
	}
       
   ///insert tabtrans utk fee user
   tabtrans_id_user:=trans_detail_2+1
	if _, err_sql = tx.Exec("insert into TABTRANS(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
		"pokok,adm,kuitansi,userid,keterangan,no_rekening_vs,kode_kantor,jam,waktu,tgl_real_trans)"+
		" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_id_user,tanggal_register,rekening_nasabah,"204","100",fee_user,
		"0",kwitansi2,"9999",desc_user,rekening_nasabah,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert tabtrans untuk user(fee_user) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	 }

	 ///insert ke transaksi_master utk fee user
	 id_master_user:=tabtrans_id_user+1
	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
        id_master_user,"TAB",kwitansi2,tanggal_register,desc_user,"TAB","139198114",
        "9999","",kode_kantor,"0",""); err_sql != nil {
		 tx.Rollback()
		 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi_master untuk user(fee_user) , desc :"+err_sql.Error())
	   return "0","Kesalahan Server",err_sql.Error()
	}
	
	///insert detail  pertama utk fee_user
	detail_trans_user_1:=id_master_user+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		detail_trans_user_1,id_master_user,perkiraan_sukarela,fee_user,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi detail debit untuk user(fee_user) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	 }
	 

	 ///insert detail 2 utk fee_user
	detail_trans_user_2:=detail_trans_user_1+1
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		detail_trans_user_2,id_master_user,perkiraan_sukarela,"0",fee_user,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi detail kredit untuk user(fee_user) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///operasi sql untuk update saldo akhir fee user
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE NO_REKENING =? ",fee_user,rekening_nasabah); 
	 err_sql != nil {
	 tx.Rollback()
	 log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat update saldo untuk user(fee_user) , desc :"+err_sql.Error())
	 return "0","Kesalahan Server",err_sql.Error()
   }

   desc_lkm:="Fee LKM "+input.Keterangan+" ke IDPEL "+input.Tujuan+" dari rekening "+rekening_nasabah

   //insert tabtrans untuk fee lkm...
   tabtrans_lkm:=detail_trans_user_2+1
   fmt.Println(tabtrans_lkm)
   if _, err_sql = tx.Exec("insert into TABTRANS(tabtrans_id,tgl_trans,no_rekening,kode_trans,my_kode_trans,"+
		"pokok,adm,kuitansi,userid,keterangan,kode_kantor,jam,waktu,tgl_real_trans)"+
		" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tabtrans_lkm,tanggal_register,rekening_nasabah,"202","200",fee_lkm,
		"0",input.Reffid,"9991",desc_lkm,kode_kantor,time,timestamp,tanggal_register); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert tabtrans untuk lkm , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

	///insert tabmaster untuk fee lkm
	tabmaster_lkm:=tabtrans_lkm+1
    fmt.Println(tabmaster_lkm)
	if _, err_sql = tx.Exec("insert into transaksi_master values(?,?,?,?,?,?,?,?,?,?,?,?)",
	tabmaster_lkm,"TAB",input.Reffid,tanggal_register,desc_lkm,"TAB","139198114",
	"9990","",kode_kantor,"0",""); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi_master untuk lkm , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}


	///insert detail trans 1 untuk fee lkm 
	detail_trans_lkm_1:=tabmaster_lkm+1
	fmt.Println(detail_trans_lkm_1)
	fmt.Println(perkiraan_sukarela)
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		detail_trans_lkm_1,tabmaster_lkm,perkiraan_sukarela,fee_lkm,"0","",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi detail debit untuk lkm , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}
	
	///insert detail trans 2 untuk fee lkm
	detail_trans_lkm_2:=detail_trans_lkm_1+1
	fmt.Println(detail_trans_lkm_2)
	fmt.Println(perkiraan_pendapatan)
	if _, err_sql = tx.Exec("insert into transaksi_detail values(?,?,?,?,?,?,?)",
		detail_trans_lkm_2,tabmaster_lkm,perkiraan_pendapatan,"0",fee_lkm,"",kode_kantor); err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat insert transaksi detail kredit untuk lkm , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
	}

  ////......update saldo perkiraan sukarela
	if _, err_sql = tx.Exec("UPDATE perkiraan SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE KODE_PERK =? ",input.Adm,perkiraan_sukarela); 
	err_sql != nil {
	tx.Rollback()
	log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan perkiraan sukarela saat untuk fee_lkm, desc :"+err_sql.Error())
	return "0","Kesalahan Server",err_sql.Error()
  }

  	///operasi sql untuk update saldo akhir fee lkm
	if _, err_sql = tx.Exec("UPDATE TABUNG SET SALDO_AKHIR = SALDO_AKHIR - ? WHERE NO_REKENING =? ",fee_lkm,rekening_nasabah); 
	err_sql != nil {
		tx.Rollback()
		log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan pada saat update saldo untuk user(fee_user) , desc :"+err_sql.Error())
		return "0","Kesalahan Server",err_sql.Error()
     }


   ////......update saldo perkiraan pendapatan
	if _, err_sql = tx.Exec("UPDATE perkiraan SET SALDO_AKHIR = SALDO_AKHIR + ? WHERE KODE_PERK =? ",input.Adm,perkiraan_pendapatan); 
	err_sql != nil {
	tx.Rollback()
	log.Println("Nasabah ID "+input.NasabahID+" gagal melakukan payment,terdapat kesalahan perkiraan pendapatan saat untuk fee_lkm, desc :"+err_sql.Error())
	return "0","Kesalahan Server",err_sql.Error()
  }

  ///lakukan commit
  err_sql=tx.Commit()
  if(err_sql!=nil){
	  tx.Rollback()
	  return "0","Kesalahan Server",err_sql.Error()
  }

   return "1","Core response oke",""
}
















