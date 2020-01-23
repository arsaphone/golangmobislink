package helpers

import(
	"strings"
	"github.com/schigh/str"
	"strconv"
	"fmt"
	 "crypto/sha1"
)

func Substr(kata string,indeks int) string{
	runes := []rune(kata)
	output:=string(runes[indeks:len(kata)])
    return output
}

func CreateTransID(maxtrans string) string {
	if(maxtrans==""){
		return "000000001"
	}
	nilai:=Substr(maxtrans,0)
	kode,_:=strconv.Atoi(nilai)
	kode_new:=kode+1
	return strconv.Itoa(kode_new)
}



func CreateNasabahID(kode_kantor string,templateNID string,nasabah_id string) (string ,string) {
    jumlah_pagar:=strings.Count(templateNID,"#")
	jumlah_9:=strings.Count(templateNID,"9")
	jumlah_kode_kantor:=len(kode_kantor)+1
	repeat_pagar:=strings.Repeat("#",jumlah_pagar)
	repeat_9:=strings.Repeat("9",jumlah_9)
	subs_kode_kantor:=Substr(nasabah_id,jumlah_kode_kantor)
	subs_kode_kantor_num,_:=strconv.Atoi(subs_kode_kantor)
	result_subs:=subs_kode_kantor_num+1
	//kurang_subs:=jumlah_9-len(strconv.Itoa(result_subs))
	kode:=str.Pad(strconv.Itoa(result_subs), "0",jumlah_9, str.PadLeft)
    fmt.Println("the code is  "+kode)
	nid:=Replace(templateNID,repeat_pagar,kode_kantor)
	nidc:=Replace(nid,repeat_9,kode)
	s:=Replace(nidc,"[","")
	as:=Replace(s,"]","")
	return as,kode
}

func FixRekening(no_rekening,kode_kantor,template string) string {
	jumlah_kode_kantor:=len(kode_kantor)
	b := str.Substring(no_rekening, jumlah_kode_kantor, 2)
	jumSb2 := jumlah_kode_kantor+2;
	jumlah_pagar:=strings.Count(template,"#")
	jumlah_9:=strings.Count(template,"9")
	jumlah_x:=strings.Count(template,"X")
	repeat_x:=strings.Repeat(template,jumlah_x)
	repeat_pagar:=strings.Repeat("#",jumlah_pagar)
	repeat_9:=strings.Repeat("9",jumlah_9)
	subs_kode_kantor:=Substr(no_rekening,jumSb2)
	nid:=Replace(template,repeat_pagar,"071")
	nidc:=Replace(nid,repeat_x,b)
	nidd:=Replace(nidc,repeat_9,subs_kode_kantor)
	s:=Replace(nidd,"[","")
	as:=Replace(s,"]","")
	return as
}


func CreateTabProgram(nasabah_id,kode_program,kode_kantor,jumlah_rek,nid string) string {
	   jumlah_pagar:=strings.Count(nid,"#")
	   jumlah_9:=strings.Count(nid,"9")
	   jumlah_x:=strings.Count(nid,"X")
	   repeat_x:=strings.Repeat("X",jumlah_x)
	   repeat_pagar:=strings.Repeat("#",jumlah_pagar)
	   repeat_9:=strings.Repeat("9",jumlah_9)
	   substr_nid:=Substr(nasabah_id,jumlah_9)
	   fmt.Println("Nasabah ID "+nasabah_id)
	   fmt.Println(substr_nid)
	   substr_nid2,_:=strconv.Atoi(substr_nid)
	   substr_nid3:=substr_nid2+0
	   fmt.Println(substr_nid3)
	   padRight:=str.Pad(jumlah_rek,strconv.Itoa(substr_nid3),jumlah_9, str.PadRight)
	   fmt.Println("PAd RIght is "+padRight)
	   nid1:=Replace(nid,repeat_pagar,kode_kantor)
	   nid2:=Replace(nid1,repeat_x,kode_program)
	   nid3:=Replace(nid2,repeat_9,padRight)
	   nid4:=Replace(nid3,"[","")
	   as:=Replace(nid4,"]","")
	   return as
}

func Replace(kata,ganti,pengganti string) string{
	if(ganti==""){
		nid:=strings.Replace(kata,ganti,pengganti,0)
		return nid
	} else {
		nid:=strings.Replace(kata,ganti,pengganti,-1)
		return nid
	}
}

func Sha1(kata string) string {
	h := sha1.New()
    h.Write([]byte(kata))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x",bs)
}
