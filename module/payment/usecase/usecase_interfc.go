package usecase

import(
	"github.com/gwlkm_service/module/payment/model"
)

type PaymentUseCase interface{
	PayHandler(model.InputPay)(string,string,string)
}