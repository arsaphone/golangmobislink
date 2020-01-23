package config

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strings"
	"os"
	"log"
)

type DB struct{
	DBE   *sqlx.DB
}

type DBSys struct{
	DB   *sqlx.DB
}

var dbConn=&DB{}
var dbsys=&DBSys{}

func Connectdb(config TomlConfig) (*DB , error) {
	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	f, err := os.OpenFile(dir+"/error_koneksi_database.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	s := []string{config.Database.USER, ":", config.Database.PASSWORD, "@tcp(", config.Database.HOST, ":",config.Database.PORT, ")/", config.Database.NAME_DB}
	 dbParam := strings.Join(s, "")
	 //db, err := sqlx.Connect("mysql", "root:@tcp(127.0.0.1)/kube71")
	db, err := sqlx.Connect("mysql",dbParam)
	//fmt.Println(dbParam)
	if err != nil {
		log.Println("Error tidak terhubung ke database ,desc error : "+err.Error())
	}
    dbConn.DBE=db
	return dbConn, err
}


func ConnectdbSys(config TomlConfig) (*DBSys , error) {
	dir, errs := os.Getwd()
	if errs != nil {
		log.Fatal(errs.Error())
	}

	f, err := os.OpenFile(dir+"/error_koneksi_database_sys.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	s := []string{config.Database_Sys.USER, ":", config.Database_Sys.PASSWORD, "@tcp(", config.Database_Sys.HOST, ":",config.Database_Sys.PORT, ")/", config.Database_Sys.NAME_DB}
	 dbParam := strings.Join(s, "")
	 //db, err := sqlx.Connect("mysql", "root:@tcp(127.0.0.1)/kube71")
	db, err := sqlx.Connect("mysql",dbParam)
	fmt.Println(dbParam)
	if err != nil {
		log.Println("Error tidak terhubung ke database sys ,desc error : "+err.Error())
	}
    dbsys.DB=db
	return dbsys, err
}
