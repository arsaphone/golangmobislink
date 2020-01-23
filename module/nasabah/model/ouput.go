package model

import (
	"database/sql"
)

type HistoryTrans struct {
	KodeRekening       string       `json:"no_rekening"      db:"NO_REKENING"`
	TabtransID         string       `json:"tabtrans_id"      db:"TABTRANS_ID"`
	TglTrans           string       `json:"tgl_trans"        db:"TGL_TRANS"`
	MyKodeTrans        string       `json:"my_kode_trans"    db:"MY_KODE_TRANS"`
	Pokok              float64      `json:"pokok"            db:"POKOK"`
	Kuitansi           string       `json:"kuitansi"         db:"KUITANSI"`
	Keterangan         string       `json:"keterangan"       db:"KETERANGAN"`
	Status             string       `json:"status"           db:"STATUS"`
}

type Produk struct{
	KodeRekening       string       `json:"no_rekening"      db:"NO_REKENING"`
	KodeProduk         string       `json:"kode_produk"      db:"KODE_PRODUK"`
	DeskripsiProduk    string       `json:"deskripsi_produk" db:"DESKRIPSI_PRODUK"`
}


type DataRekening struct{
	NamaNasabah  string          `json:"nama_nasabah"   db:"NAMA_NASABAH"`
	NoRekening   string          `json:"no_rekening"   db:"NO_REKENING"`
	SaldoAkhir   sql.NullFloat64 `json:"saldo_akhir"   db:"SALDO_AKHIR"`
 }

 type DataNasabah struct{
    Nasabah_id   string   
    Phone_number string    
  }

  type TabProduk struct{
    KodeProduk       string     `json:"kode_produk"          db:"KODE_PRODUK"`
  	SukuBungaDefault string     `json:"suku_bunga_default"   db:"SUKU_BUNGA_DEFAULT"`
	SetoranPertama   string     `json:"setoran_pertama"      db:"SETORAN_PERTAMA"`
	Jenis            string     `json:"jenis"                db:"JENIS"`
}

	type DetailTabungan struct{
		Nasabah_id string
		Deskripsi_produk string
		No_rekening string
		Nama_nasabah string
		Saldo string
	}
	

	type TabNasabah struct{
     Nasabah_id       string                      `json:"NASABAH_ID"           db:"NASABAH_ID"`
  	 NoRekening       string                      `json:"NO_REKENING"          db:"NO_REKENING"`
	 KodeProduk       string                      `json:"KODE_PRODUK"          db:"KODE_PRODUK"`
	 DeskripsiProduk  string                      `json:"DESKRIPSI_PRODUK"     db:"DESKRIPSI_PRODUK"`
	 SaldoAkhir       sql.NullFloat64             `json:"SALDO_AKHIR"          db:"SALDO_AKHIR"`
	}
	
	type NasabahPoin struct{
		Nasabah_id       string     `json:"NASABAH_ID"           db:"NASABAH_ID"`
  	    NamaNasabah      string     `json:"NAMA_NASABAH"         db:"NAMA_NASABAH"`
		Telpon           string     `json:"TELPON"               db:"TELPON"`
		TotalPoin        string     `json:"TOTAL_POIN"           db:"TOTAL_POINT"`
	}

	type PayProductMapping struct{
		Code             string     `json:"CODE"                 db:"CODE"`
     	Kode_Trans       string     `json:"KODE_TRANS"           db:"KODE_TRANS"`
		Deskripsi        string     `json:"DESKRIPSI"            db:"DESKRIPSI"`
		Type             string     `json:"TYPE"                 db:"TYPE"`
	}

	type ResultTransfer struct{
		NamaPengirim        string     `json:"nama_pengirim"`
  	    NamaPenerima        string     `json:"nama_penerima"`
		Pengirim            string     `json:"pengirim"`
  	    Penerima            string     `json:"penerima"`
		Waktu               string     `json:"waktu"`
		Nominal             float64    `json:"status"`
	}


	type InfoNasabah struct{
		NamaNasabah             string     `db:"NAMA_NASABAH"`
		Alamat                  string     `db:"ALAMAT"`
		JenisKelamin            string     `db:"JENIS_KELAMIN"`
		TempatLahir             string     `db:"TEMPATLAHIR"`
		TanggalLahir            string     `db:"TGLLAHIR"`
		TglRegister             string     `db:"TGL_REGISTER"`
		Email                   string     `db:"EMAIL"`
		TotalPoint              string     `db:"TOTAL_POINT"`
	}





