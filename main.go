package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jamesonhm/gochain/internal/tasty"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// defer functions are processed LIFO, context cancel must run before scheduler shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Env Variable Load
	godotenv.Load()
	tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastySB)
	login := tasty.LoginInfo{
		Login:      mustEnv("TASTY_USER"),
		Password:   mustEnv("SB_PASSWORD"),
		RememberMe: true,
	}
	err := tastyClient.CreateSession(ctx, login)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Tasty Session", slog.String("error creating session", err.Error()))
	}
	logger.Info("Tasty Session", "tasty user", tastyClient.GetUser())

	customer, err := tastyClient.GetCustomer(ctx)
	fmt.Println(customer)

	accts, err := tastyClient.GetAccounts(ctx)
	fmt.Printf("%+v\n", accts)

	acctParams := tasty.AcctNumParams{
		AcctNum: accts[0].Account.AccountNumber,
	}

	acct, err := tastyClient.GetAccount(ctx, &acctParams)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", acct)

	status, err := tastyClient.GetAccountTradingStatus(ctx, &acctParams)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", status)

	posParams := tasty.AccountPositionParams{
		AccountNumber: accts[0].Account.AccountNumber,
	}
	pos, err := tastyClient.GetAccountPositions(ctx, &posParams)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", pos)

	bal, err := tastyClient.GetAccountBalances(ctx, &acctParams)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", bal)

	act := true
	syms := []string{"XSP 250423P00529000"}
	sym := "XSP 250423C00529000"
	expd := false
	eqOpParams := tasty.EquityOptionsParams{
		Active:      &act,
		Symbol:      &syms,
		WithExpired: &expd,
	}
	eqOpts, err := tastyClient.GetEquityOptions(ctx, &eqOpParams)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", eqOpts)

	eqOpSym := tasty.EquityOptionSymbol{
		Symbol: sym,
		Active: &act,
	}
	eqOpt, err := tastyClient.GetEquityOption(ctx, &eqOpSym)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", eqOpt)
	//bt, err := tastyClient.BacktestSession(ctx)
	//if err != nil {
	//	logger.LogAttrs(ctx, slog.LevelError, "Tasty Backtest", slog.String("error creating session", err.Error()))
	//} else {
	//	fmt.Println("backtest token", *bt.Token)
	//}
	//tastyClient.CancelBacktestSession(ctx)

}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Env variable %s required", key))
	}
	return val
}
