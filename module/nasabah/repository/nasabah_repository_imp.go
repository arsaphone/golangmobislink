package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	_ "strings"
	_ "time"

	"github.com/gwlkm_service/module/nasabah/model"
	"github.com/jmoiron/sqlx"
)

type MySQLNasabahRepo struct {
	Conn    *sqlx.DB
	ConnSys *sqlx.DB
}

func NewSQLNasabahRepo(Conn *sqlx.DB, ConnSys *sqlx.DB) NasabahRepo {
	return &MySQLNasabahRepo{
		Conn:    Conn,
		ConnSys: ConnSys,
	}
}

func (m *MySQLNasabahRepo) GetEmailNasabah(nasabah_id string) string {
	var email string
	m.Conn.Get(&email, "select EMAIL from nasabah where NASABAH_ID=?", nasabah_id)
	return email
}

func (m *MySQLNasabahRepo) CheckValidNasabahID(nasabah_id string) error {
	var nasabah string
	err := m.Conn.Get(&nasabah, "select nasabah_id from reg_pay_account where NASABAH_ID=?", nasabah_id)
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) CheckNoTelp(no string) error {
	var nasabah_id string
	err := m.Conn.Get(&nasabah_id, "select NASABAH_ID from nasabah where TELPON=?", no)
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) GetNamaNasabah(nasabah_id string) string {
	var nama string
	m.Conn.Get(&nama, "select NAMA_NASABAH from nasabah where NASABAH_ID=?", nasabah_id)
	return nama
}

func (m *MySQLNasabahRepo) CheckNoNIK(nik string) error {
	var nasabah_id string
	err := m.Conn.Get(&nasabah_id, "select NASABAH_ID from nasabah where NO_ID=?", nik)
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) GetTemplateNID(nid string) string {
	var keyvalue string
	err := m.ConnSys.Get(&keyvalue, "SELECT KEYVALUE FROM sys_mysysid WHERE KEYNAME=?", nid)
	if err != nil {
		fmt.Println(err.Error())
	}
	//fmt.Println(m.ConnSys)
	//fmt.Println(m.Conn)
	return keyvalue
}

func (m *MySQLNasabahRepo) GetKodeKantor() string {
	var kode_kantor string
	m.Conn.Get(&kode_kantor, "SELECT KODE_KANTOR FROM app_kode_kantor")
	return kode_kantor
}

func (m *MySQLNasabahRepo) GetTabProduk() []*model.TabProduk {
	list_produk := make([]*model.TabProduk, 0)
	m.Conn.Select(&list_produk, "SELECT KODE_PRODUK,SUKU_BUNGA_DEFAULT,SETORAN_PERTAMA,JENIS "+
		" FROM tab_produk WHERE automatic_create LIKE '1'")
	return list_produk
}
func (m *MySQLNasabahRepo) GetListTransaksi(no_rekening string, limit, offset int) []*model.HistoryTrans {
	list_trans := make([]*model.HistoryTrans, 0)
	err := m.Conn.Select(&list_trans, "SELECT NO_REKENING,TABTRANS_ID,TGL_TRANS,MY_KODE_TRANS,POKOK,KUITANSI,KETERANGAN,CASE"+
		" WHEN MY_KODE_TRANS='100' THEN 'Pemasukkan'"+
		" ELSE 'Pengeluaran'"+
		" END AS STATUS"+
		" FROM TABTRANS where NO_REKENING = ? order by TABTRANS_ID  DESC  limit ? offset ?  ", no_rekening, limit, offset)
	if err != nil {
		fmt.Println(err.Error())
	}
	///	fmt.Println(err.Error())
	return list_trans
}

func (m *MySQLNasabahRepo) GetOneTabProduk() model.TabProduk {
	var tab_produk model.TabProduk
	m.Conn.Get(&tab_produk, "SELECT KODE_PRODUK,SUKU_BUNGA_DEFAULT,SETORAN_PERTAMA"+
		" FROM tab_produk WHERE automatic_create LIKE '1'")
	return tab_produk
}
func (m *MySQLNasabahRepo) UpdateStatusAgenNasabah(nasabah_id string) error {
	tx, _ := m.Conn.Begin()
	_, err := tx.Exec("UPDATE nasabah SET  agen_id = ? WHERE nasabah_id = ?", nasabah_id, nasabah_id)
	tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) GetNasabahTabPayment(nasabah_id, kode_produk string) (string, sql.NullFloat64, error) {
	type Nasabah struct {
		No_Rekening string          `db:"no_rekening"`
		Saldo_akhir sql.NullFloat64 `db:"saldo_akhir"`
	}
	var nasabah Nasabah
	err := m.Conn.Get(&nasabah, "SELECT no_rekening,saldo_akhir FROM TABUNG WHERE nasabah_id=? and kode_produk=?", nasabah_id, kode_produk)
	if err != nil {
		return "", nasabah.Saldo_akhir, err
	}
	return nasabah.No_Rekening, nasabah.Saldo_akhir, err
}

func (m *MySQLNasabahRepo) GetPinByNasabahId(nasabah_id string) (error, string) {
	var PinNasabah string
	err := m.Conn.Get(&PinNasabah, "select password from register_pay_service where nasabah_id=?", nasabah_id)
	if err != nil {
		return err, ""
	}
	return nil, PinNasabah
}

func (m *MySQLNasabahRepo) UpdatePinNasabah(nasabah_id string, pin_baru string) error {
	tx, _ := m.Conn.Begin()
	_, err := tx.Exec("UPDATE register_pay_service SET password = ? WHERE nasabah_id = ?", pin_baru, nasabah_id)
	tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) GetPayProductMapping(code string) model.PayProductMapping {
	var model model.PayProductMapping
	err := m.Conn.Get(&model, "SELECT CODE,KODE_TRANS,DESKRIPSI,TYPE from pay_product_mapping where code=?", code)
	if err != nil {
		fmt.Println(err.Error())
	}

	return model
}

func (m *MySQLNasabahRepo) GetPoinByNasabahId(nasabah_id string) (error, model.NasabahPoin) {
	var poin model.NasabahPoin
	err := m.Conn.Get(&poin, "SELECT NASABAH_ID, NAMA_NASABAH, TELPON,TOTAL_POINT FROM nasabah where nasabah_id=?", nasabah_id)
	if err != nil {
		return err, poin
	}
	return nil, poin

}

func (m *MySQLNasabahRepo) GetCountTransaksi(no_rekening string) int {
	var jumlah int
	m.Conn.Get(&jumlah, "SELECT count(TABTRANS_ID) from TABTRANS where no_rekening=?", no_rekening)
	return jumlah
}

func (m *MySQLNasabahRepo) GetTabNasabah(nasabah_id string) (error, []*model.TabNasabah) {
	list_tab := make([]*model.TabNasabah, 0)
	err := m.Conn.Select(&list_tab, "SELECT pl.NASABAH_ID, pl.NO_REKENING, pl.KODE_PRODUK,"+
		"pn.DESKRIPSI_PRODUK"+
		",pl.SALDO_AKHIR FROM TABUNG pl, tab_produk pn WHERE pl.KODE_PRODUK = pn.KODE_PRODUK AND pl.NASABAH_ID = ?", nasabah_id)
	if err != nil {
		return err, nil
	}
	return nil, list_tab
}

func (m *MySQLNasabahRepo) GetMaxNasabahId() string {
	var nasabah_id string
	m.Conn.Get(&nasabah_id, "SELECT max(nasabah_id) as hole FROM nasabah")
	return nasabah_id
}

func (m *MySQLNasabahRepo) GetMaxTransID() string {
	var trans_id string
	trans_id = ""
	m.Conn.Get(&trans_id, "SELECT max(trans_id) as hole from transaksi_detail")
	return trans_id
}

func (m *MySQLNasabahRepo) CheckNoTelponLogin(no_telp string) (string, string, error) {
	type Nasabah struct {
		Nasabah_id   string `db:"nasabah_id"`
		Phone_number string `db:"phone_number"`
	}

	var nasabah Nasabah
	err := m.Conn.Get(&nasabah, "SELECT nasabah_id,phone_number FROM reg_pay_account WHERE phone_number=?", no_telp)
	if err != nil {
		return "", "", err
	}
	return nasabah.Nasabah_id, nasabah.Phone_number, nil
}

func (m *MySQLNasabahRepo) CheckValidNoRekeningNasabah(no_rek string) error {
	var no_rek2 string
	err := m.Conn.Get(&no_rek2, "SELECT NO_REKENING from TABUNG WHERE NO_REKENING=?", no_rek)
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) GenerateNasabahID() string {
	var count int
	m.Conn.Get(&count, "select count(trans_id) as total from transaksi_detail")
	count = count + 1
	return strconv.Itoa(count)
}

func (m *MySQLNasabahRepo) GetDeskripsiTabProduk(kode_produk string) string {
	var deskripsi string
	m.Conn.Get(&deskripsi, "SELECT DESKRIPSI_PRODUK FROM TAB_PRODUK WHERE KODE_PRODUK=?", kode_produk)
	return deskripsi
}

func (m *MySQLNasabahRepo) GetKodeTabProduk(no_rekening string) string {
	var kode_produk string
	m.Conn.Get(&kode_produk, "SELECT KODE_PRODUK FROM TABUNG WHERE NO_REKENING =?", no_rekening)
	return kode_produk
}

func (m *MySQLNasabahRepo) GetJumlahNasabah() int {
	var jumlah int
	m.Conn.Get(&jumlah, "SELECT count(nasabah_id) from nasabah")
	return jumlah
}

func (m *MySQLNasabahRepo) GetJumlahRegPay() int {
	var jumlah int
	m.Conn.Get(&jumlah, "SELECT count(nasabah_id) from reg_pay_account")
	return jumlah
}

func (m *MySQLNasabahRepo) CheckSaldoByNoRekening(no_rekening string) (string, sql.NullFloat64) {
	type Nasabah struct {
		Nasabah_id  string          `db:"nasabah_id"`
		Saldo_akhir sql.NullFloat64 `db:"saldo_akhir"`
	}

	var nasabah Nasabah
	m.Conn.Get(&nasabah, "SELECT nasabah_id,saldo_akhir FROM TABUNG WHERE NO_REKENING=?", no_rekening)
	return nasabah.Nasabah_id, nasabah.Saldo_akhir
}

func (m *MySQLNasabahRepo) GetRegPayAccountByNasabahID(nasabah_id string) (string, string, error) {
	type Nasabah struct {
		Nasabah_id   string
		Phone_number string
	}
	var nasabah Nasabah
	err := m.Conn.Get(&nasabah, "SELECT nasabah_id,phone_number FROM reg_pay_account WHERE nasabah_id=?", nasabah_id)
	if err != nil {
		return "", "", err
	}
	return nasabah.Nasabah_id, nasabah.Phone_number, nil
}

func (m *MySQLNasabahRepo) CheckRegPayService(nasabah_id string, pin string) error {
	var nasabah_id_x string
	err := m.Conn.Get(&nasabah_id_x, "select NASABAH_ID from register_pay_service where NASABAH_ID=? and password=?", nasabah_id, pin)
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLNasabahRepo) CheckStatusAgenNasabah(nasabah_id string) string {
	var agen_id string
	agen_id = ""
	err := m.Conn.Get(&agen_id, "SELECT agen_id FROM nasabah WHERE NASABAH_ID=?", nasabah_id)
	if err != nil {
		return agen_id
	}
	return agen_id
}

func (m *MySQLNasabahRepo) GetTemplateRekening() string {
	var keyvalue string
	err := m.ConnSys.Get(&keyvalue, "SELECT KEYVALUE FROM sys_mysysid WHERE KEYNAME=?", "TAB_TEMPLATE_NO_REKENING")
	if err != nil {
		fmt.Println(err.Error())
	}
	return keyvalue
}

func (m *MySQLNasabahRepo) GetInfoNasabah(nasabah_id string) model.InfoNasabah {
	var info model.InfoNasabah
	err := m.Conn.Get(&info, "SELECT NAMA_NASABAH,ALAMAT,JENIS_KELAMIN,TEMPATLAHIR,TGLLAHIR,TGL_REGISTER,EMAIL,TOTAL_POINT from nasabah where NASABAH_ID = ?", nasabah_id)
	if err != nil {
		fmt.Println(err.Error())
	}
	return info
}
