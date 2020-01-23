package model

type TabProduk struct{
    KodeProduk       string     `json:"kode_produk"          db:"KODE_PRODUK"`
  	DeskripsiProduk  string     `json:"deskripsi_produk"     db:"DESKRIPSI_PRODUK"`
  	SetoranPertama   string     `json:"setoran_pertama"      db:"SETORAN_PERTAMA"`
}


type ListAnggota struct{
    NasabahID       string     `json:"nasabah_id"          db:"NASABAH_ID"`
  	NamaNasabah     string     `json:"nama_nasabah"        db:"NAMA_NASABAH"`
}