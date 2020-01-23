package repository

import(
	"github.com/gwlkm_service/module/geolokasi/model"
	"github.com/jmoiron/sqlx"
)


type MySQLGeoRepo struct {
	Conn    *sqlx.DB
}

func NewSQLGeoRepo(Conn *sqlx.DB) GeolokasiRepo {
	return &MySQLGeoRepo{
		Conn    : Conn,
	}
}

func(m  *MySQLGeoRepo)CheckNasabah(nasabah_id string)(error){
	var nasabah_id_2 string
	err:=m.Conn.Get(&nasabah_id_2,"SELECT nasabah_id from tempat_kantor where nasabah_id=?",nasabah_id)
	if(err!=nil){
		return err
	}
	return nil
}


func(m  *MySQLGeoRepo)LoadTempatNasabah()[]*model.OutputLokasi{
	list_nasabah:=make([]*model.OutputLokasi,0)
	m.Conn.Select(&list_nasabah,"SELECT no,nasabah_id,nama_nasabah,lat,lng,keterangan from tempat_kantor")
	return list_nasabah
}

func(m  *MySQLGeoRepo)UpdateTempatNasabah(model model.InputLokasi)(error){
	tx,_:= m.Conn.Begin()
	_,err:=tx.Exec("UPDATE tempat_kantor lat=?,lng=?,nama_nasabah=?,keterangan=? where nasabah_id=?",model.Lat,model.Lng,model.Nama_Nasabah,model.Keterangan,model.Nasabah_ID)
	err=tx.Commit()
    if(err!=nil){
	  return err
	}
	return nil
}

func(m  *MySQLGeoRepo)InsertTempatNasabah(model model.InputLokasi)(error){
	tx,_:= m.Conn.Begin()
	_,err:=tx.Exec("insert into tempat_kantor values(?,?,?,?,?,?)","",model.Nasabah_ID,model.Nama_Nasabah,model.Lat,model.Lng,model.Keterangan)
	err=tx.Commit()
    if(err!=nil){
	  return err
	}
	return nil
}




