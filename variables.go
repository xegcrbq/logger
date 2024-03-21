package logger

var SENDPERIOD = 5
var CALLERCOUNT = 3

type Tag string

const (
	Sql            = Tag("sql")
	General        = Tag("general")
	Service        = Tag("service")
	Warning        = Tag("warning")
	OrdersOffers   = Tag("ordersOffers")
	Balance        = Tag("balance")
	FatalPanic     = Tag("fatalPanic")
	IncomeWithdraw = Tag("incomeWithdraw")
)
