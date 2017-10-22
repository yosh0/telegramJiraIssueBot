package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"regexp"
	"encoding/json"
	"golang.org/x/net/context"
	"github.com/bot-api/telegram"
	"github.com/jasonlvhit/gocron"
	"github.com/bot-api/telegram/telebot"
)

var (
	C = Config{}
	DB_DRIVER = "postgres"
	DB_TBL = "telegram_issue_users"
	listSavedUsers = make(map[int64]SavedUser)
	keyMatchRegex = regexp.MustCompile(`\"(\w+)\":`)
	wordBarrierRegex = regexp.MustCompile(`(\w)([A-Z])`)
)

func tasks() {
	gocron.Every(1).Day().At("00:01").Do(ClearCounters)
	<- gocron.Start()
	s := gocron.NewScheduler()
	<- s.Start()
}

func ClearCounters() {
	LogFuncStr(fName(), "####################### Clear Users #######################")
	for _, SU := range listSavedUsers {
		LogFuncStr(fName(), fmt.Sprintf("Clear User %s", SU.PhoneNumber))
		SU.IssueCreatedCount = 0
		SU.clearSavedUsers()
	}
	LogFuncStr(fName(), "####################### Clear Users #######################")
}

func main() {
	go tasks()
	token := flag.String("token", C.Serv.Token, "telegram bot token")
	debug := flag.Bool("debug", C.Serv.Debug, "show debug information")
	flag.Parse()
	if *token == "" {
		log.Fatal("token flag is required")
	}

	api := telegram.New(*token)
	api.Debug(*debug)
	bot := telebot.NewWithAPI(api)
	bot.Use(telebot.Recover()) // recover if handler panic
	netCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot.HandleFunc(func(ctx context.Context) error {
		update := telebot.GetUpdate(ctx)
		if update.Message == nil {
			return nil
		}
		err := initCustomKeyboard(update, api, ctx)
		if err != nil {
			return err
		}
		return nil
	})

	bot.Use(telebot.Commands(map[string]telebot.Commander{
		"start": telebot.CommandFunc(
			func(ctx context.Context, arg string) error {
				api := telebot.GetAPI(ctx)
				update := telebot.GetUpdate(ctx)
				msg := MessageHTML(update, C.StartMsg.Command, C.KbBtnText.Auth, false)
				_, err := api.Send(ctx, msg)
				msg = MessageHTML(update, C.StartMsg.Contact, C.KbBtnText.Auth, true)
				_, err = api.Send(ctx, msg)
				return err
			}),
	}))
	err := bot.Serve(netCtx)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	c, err := os.Open(".config.json")
	if err != nil {
		LogFuncStr(fName(), err.Error())
	}
	decoder := json.NewDecoder(c)
	conf := Config{}
	err = decoder.Decode(&conf)
	C = conf
}
