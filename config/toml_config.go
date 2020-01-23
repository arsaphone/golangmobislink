package config

type TomlConfig struct{
	Title             string  
	Database          database
	Database_Sys      database_sys  
	SET_ID            set_id  
	Server            server  
	Var_regagtbyagen  var_regagtbyagen
	Point_trans       point_trans
	Fee_pay_pulsa     fee_pay_pulsa
	Perkiraan         perkiraan
}

type database struct{
	HOST        string 
	PORT        string 
	NAME_DB     string 
	USER        string 
	PASSWORD    string 
}

type database_sys struct{
	HOST        string 
	PORT        string 
	NAME_DB     string 
	USER        string 
	PASSWORD    string 
}

type set_id struct{
	KEYNAME_NASABAH_ID string
	KEYNAME_NOREK      string
}

type var_regagtbyagen struct{
	BIAYA_PENDAFTARAN float64
	BIAYA_FEE_ADM     float64
	BIAYA_FEE_AGEN    float64
	BIAYA_DAFTAR_AGEN float64
}

type point_trans struct{
	POINT_TRANS_PULSA         int
	POINT_REG_AGT_BYAGENT     int
	POINT_REK_BERJANGKA_UMROH int
}

type fee_pay_pulsa struct{
	FEE_USER        int
    FEE_LKM         int
    FEE_AGEN_USER   int
}

type perkiraan struct{
	SUKARELA_LKM               string
    DEPOSIT_ECHANNEL           string
    PENDAPATAN_NON_OPERASIONAL string
}


type server struct{
	Port string    
}