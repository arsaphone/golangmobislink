package model

type OutputLokasi struct{
	No             string        `DB:"no"`
	Nasabah_ID     string        `DB:"nasabah_id"`
    NamaNasabah    string        `DB:"nama_nasabah"`
	Lat            string        `DB:"lat"`
	Lng            string        `DB:"lng"`
	Keterangan     string        `DB:"keterangan"`
}