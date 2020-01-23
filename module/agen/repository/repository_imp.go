package repository

import(
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/gwlkm_service/module/agen/model"
)

type MySQLAgenRepo struct {
	Conn    *sqlx.DB
	ConnSys *sqlx.DB
}

func NewSQLAgenRepo(Conn *sqlx.DB,ConnSys *sqlx.DB) AgenRepo {
	return &MySQLAgenRepo{
		Conn    : Conn,
		ConnSys : ConnSys,
	}
}

func(m  *MySQLAgenRepo)GetCountTabungan(nasabah_id string)(int){
	var count int
	m.Conn.Get(&count,"SELECT count(no_rekening) from tabung where nasabah_id=?",nasabah_id)
	return count
}


func(m  *MySQLAgenRepo)CheckNasabahID(nasabah_id string)(error){
	var nasabah_id_2 string
	err:=m.Conn.Get(&nasabah_id_2,"SELECT nasabah_id from tabung where nasabah_id=?",nasabah_id)
	if(err!=nil){
		return err
	}
   return nil
}

func(m  *MySQLAgenRepo)GetBiayaAdminTabungan(keyname,kode_program string)(string){
  var biaya  string	
  m.ConnSys.Get(&biaya,"SELECT keyvalue from sys_mysysid WHERE keyname = ?",keyname)
  return biaya
}



func(m  *MySQLAgenRepo)GetBiayaTabProgram(kode_produk string)(string,float64,float64){
	type BiayaTab struct{
		Deskripsi_Produk  string             `db:"DESKRIPSI_PRODUK"`
	    SukuBunga         sql.NullFloat64    `db:"SUKU_BUNGA_DEFAULT"`
		Biaya_Tab_Program sql.NullFloat64    `db:"BIAYA_TAB_PROGRAM"`
	}

	var biaya BiayaTab
	m.Conn.Get(&biaya,"SELECT DESKRIPSI_PRODUK,SUKU_BUNGA_DEFAULT,BIAYA_TAB_PROGRAM FROM TAB_PRODUK WHERE kode_produk = ?",kode_produk)
	return biaya.Deskripsi_Produk,biaya.SukuBunga.Float64,biaya.Biaya_Tab_Program.Float64
}


func(m  *MySQLAgenRepo) GetReferalID(nasabah_id string)(string,error){
	var referal_id string
	err:=m.Conn.Get(&referal_id,"SELECT referal_id FROM nasabah WHERE NASABAH_ID = ?",nasabah_id)
	if(err!=nil){
		return "tidak_ada",err
	}
	return referal_id,nil
}


func(m  *MySQLAgenRepo)GetAgenKeyValue(keyvalue string) float64 {
	var keyvalue2 sql.NullFloat64
	err:=m.ConnSys.Get(&keyvalue2,"SELECT KEYVALUE FROM sys_mysysid WHERE KEYNAME=?",keyvalue)
	if(err!=nil){
		fmt.Println(err.Error())
	}
	return keyvalue2.Float64
}

func(m  *MySQLAgenRepo) GetTabProgram() []*model.TabProduk {
	list_produk:=make([]*model.TabProduk,0)
	m.Conn.Select(&list_produk,"SELECT KODE_PRODUK,DESKRIPSI_PRODUK,SETORAN_PERTAMA"+ 
	" FROM tab_produk WHERE JENIS='SSS'")
	return list_produk
}

func(m  *MySQLAgenRepo)GetListAnggota(nasabah_id string)([]*model.ListAnggota){
	list_anggota:=make([]*model.ListAnggota,0)
	m.Conn.Select(&list_anggota,"select NASABAH_ID,NAMA_NASABAH from NASABAH where referal_id=?",nasabah_id)
	return list_anggota
}


func(m  *MySQLAgenRepo)GetNoRekeningByJenisTabungan(nasabah_id,jenis string) string {
	var no_rekening string
	err:=m.Conn.Get(&no_rekening,"SELECT t.NO_REKENING"+ 
	 " FROM tabung t join tab_produk p on t.kode_produk=p.kode_produk"+
	 " WHERE t.nasabah_id =? and p.jenis=?",nasabah_id,jenis)
	 if(err!=nil){
		 fmt.Println(err.Error())
	 }
   return no_rekening
}

func(m  *MySQLAgenRepo)GetSetoranPertamaJenisTabungan(jenis_tabungan string) float64 {
	var hasil_sum sql.NullFloat64
	err:=m.Conn.Get(&hasil_sum,"SELECT SETORAN_PERTAMA AS HOLE FROM tab_produk WHERE JENIS =?",jenis_tabungan)
	if(err!=nil){
		fmt.Println(err.Error())
	}
	return hasil_sum.Float64
}




func(m  *MySQLAgenRepo)GetSumTabungan() float64 {
	var hasil_sum sql.NullFloat64
	err:=m.Conn.Get(&hasil_sum,"SELECT SUM(SETORAN_PERTAMA) AS HOLE FROM tab_produk WHERE automatic_create =?","1")
	if(err!=nil){
		fmt.Println(err.Error())
	}
	return hasil_sum.Float64
}


func(m  *MySQLAgenRepo)GetSaldoAkhirPerkiraan(kode_perkiraan string) float64 {
	var saldo_akhir sql.NullFloat64
	err:=m.Conn.Get(&saldo_akhir,"SELECT SALDO_AKHIR FROM perkiraan WHERE KODE_PERK =?",kode_perkiraan)
	if(err!=nil){
		fmt.Println(err.Error())
	}
	return saldo_akhir.Float64
}


