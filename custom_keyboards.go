package main

import (
	"fmt"
	"time"
	"errors"
	"regexp"
	"strconv"
	"database/sql"
	"unicode/utf8"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"github.com/bot-api/telegram"
	"github.com/Masterminds/squirrel"
)

func NewCustomKeyboard(text string, request_contact bool) [][]telegram.KeyboardButton {
	r := make([][]telegram.KeyboardButton, 1)
	r[0] = []telegram.KeyboardButton{
		{
			Text: text,
			RequestContact: request_contact,
		},
	}
	return r
}

func MessageHTML(update *telegram.Update, text string, kbtext string, rc bool) telegram.MessageCfg {
	textHtml := TextHTMLBold(text)
	msg := telegram.NewMessage(update.Chat().ID, textHtml)
	msg.ParseMode = "HTML"
	if kbtext == "" {
		return msg
	}
	msg.ReplyMarkup = telegram.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: NewCustomKeyboard(kbtext, rc),
	}
	return msg
}

func TextHTMLBold(text string) string {
	return fmt.Sprintf("<b>%s</b>", text)
}

func ValidateByPhoneNumber(update *telegram.Update) (SavedUser, error) {
	var SU SavedUser
	rex, err := regexp.Compile(`^\+?(7\d{10})$`)
	if err != nil {
		LogFuncStr(fName(), err.Error())
		return SU, err
	}
	res := rex.FindStringSubmatch(update.Message.Contact.PhoneNumber)
	if res != nil {
		LogFuncStr(fName(), res[0])
		LogFuncStr(fName(), res[1])
		connect, err := sql.Open(C.DB.Type, dbInit())
		if err != nil {
			LogFuncStr(fName(), err.Error())
			return SU, err
		}
		cache := squirrel.NewStmtCacheProxy(connect)
		tiu := NewTelegramIssueUsers(cache)
		tiu.PhoneNumber = res[1]
		if err := tiu.LoadByPhoneNumber(); err != nil {
			LogFuncStr(fName(), err.Error())
			return SU, err
		}
		if tiu != nil {
			if update.Message.From.ID != tiu.UserId {
				if tiu.UserId == 0 {
					tiu.UserId = update.Message.From.ID
				} else {
					err := errors.New("NUMBER OR ID INCORRECT")
					LogFuncStr(fName(), err.Error())
					return SU, err
				}
			} else {
				LogFuncStr(fName(), "update.Message.From.ID == tiu.UserId")
			}
			tiu.UserName = fmt.Sprintf("%s %s %s", update.Message.From.FirstName, update.Message.From.Username, update.Message.From.LastName)
			tiu.UpdatedAt = time.Now().Unix()
			tiu.Update()

			SU = NewSavedUser(update.Message.From.ID, tiu)
			SU.TgID = update.Message.From.ID
			SU.FirstName = tiu.FirstName
			SU.MiddleName = tiu.MiddleName
			SU.LastName = tiu.LastName
			SU.PhoneNumber = tiu.PhoneNumber
			return SU, nil
		}
		connect.Close()
	}
	err = errors.New("Phone error")
	return SU, err
}

func ContactOwner(update *telegram.Update) bool {
	if update.Message.Contact != nil {
		if update.Message.Contact.UserID == update.Message.From.ID {
			return true
		}
	}
	return false
}

func RemoveAuth(update *telegram.Update) bool {
	if _, ok := listSavedUsers[update.Message.From.ID]; ok {
		LogFuncStr(fName(), "Remove Auth from memory")
		LogSavedUsers(fName(), listSavedUsers[update.Message.From.ID])
		delete(listSavedUsers, update.Message.From.ID)
		return true
	}
	LogFuncStr(fName(), "Auth Not found in memory")
	return false
}

func initCustomKeyboard(update *telegram.Update, api *telegram.API, ctx context.Context) error {
	if update.Message.Contact != nil {
		if !ContactOwner(update) {
			LogFuncStr(fName(), "Send Not Own Contact Not Allowed")
			if !RemoveAuth(update) {

			}
			msg := MessageHTML(update, C.BotMessage.ContactNotOwn, C.KbBtnText.Auth, true)
			_, err := api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}
			err = errors.New(C.BotMessage.ContactNotOwn)
			return err
		}
		SU, err := ValidateByPhoneNumber(update)
		if err != nil {
			LogFuncStr(fName(), err.Error())
			LogFuncStr(fName(), err.Error())
			msg := MessageHTML(update, C.BotMessage.NeedAuth, C.KbBtnText.Auth, true)
			_, err := api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}
			return err
		}
		SU.updateSavedUsers()
	}

	LogFuncStr(fName(), "################# ALL MESSAGES #################")
	LogFuncStr(fName(), update.Message.Text)
	LogFuncStr(fName(), "################# ALL MESSAGES #################")

	SU, err := inSavedUsersList(update.Message.From.ID)
	if err != nil {
		LogFuncStr(fName(), err.Error())
		msg := MessageHTML(update, C.BotMessage.NeedAuth, C.KbBtnText.Auth, true)
		_, err := api.Send(ctx, msg)
		if err != nil {
			LogFuncStr(fName(), err.Error())
			return err
		}
	} else {


		connect, err := sql.Open(C.DB.Type, dbInit())
		if err != nil {
			LogFuncStr(fName(), err.Error())
			return err
		}
		cache := squirrel.NewStmtCacheProxy(connect)
		tiu := NewTelegramIssueUsers(cache)
		tiu.PhoneNumber = SU.PhoneNumber
		if err := tiu.LoadByPhoneNumber(); err != nil {
			LogFuncStr(fName(), err.Error())
			return err
		}
		tiu.UpdatedAt = time.Now().Unix()
		tiu.Update()
		connect.Close()


		if SU.IssueCreatedCount >= C.JiraIssue.Limit {
			msg := MessageHTML(update, C.BotMessage.IssueLimit, "", false)
			_, err = api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}
			err = errors.New(C.BotMessage.IssueLimit)
			return err
		}
		if SU.Messages.FinishMessage == true && SU.IssueCreated == true {
			LogFuncStr(fName(), "TicketCreationTimeout1")
			timeout, err := SU.isTicketCreationTimeout()
			if err != nil {
				LogFuncStr(fName(), err.Error())
			}
			if timeout == "" {
				err = errors.New("Timeout is nil")
				LogFuncStr(fName(), err.Error())

				msg := MessageHTML(update, C.BotMessage.IssueAbout, "", false)
				_, err = api.Send(ctx, msg)
				if err != nil {
					LogFuncStr(fName(), err.Error())
					return err
				}
				return err
			}
			LogFuncStr(fName(), timeout)
			msg := MessageHTML(update, fmt.Sprintf("%s %s", C.BotMessage.IssueTimeout, timeout), "", false)
			_, err = api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}
			err = errors.New(C.BotMessage.IssueTimeout)
			return err
		}
		if SU.Messages.CreationMessage == true {
			LogFuncStr(fName(), "-----------------MESSAGE LEN-----------------")
			LogFuncStr(fName(), fmt.Sprintf("Rune count: %s", strconv.Itoa(utf8.RuneCountInString(update.Message.Text))))
			LogFuncStr(fName(), fmt.Sprintf("Char count: %s", strconv.Itoa(len(update.Message.Text))))
			LogFuncStr(fName(), "-----------------MESSAGE LEN-----------------")
			if utf8.RuneCountInString(update.Message.Text) < C.JiraIssue.DescLen {
				LogFuncStr(fName(), update.Message.Text)
				msg := MessageHTML(update, C.BotMessage.IssueLess, "", false)
				_, err = api.Send(ctx, msg)
				if err != nil {
					LogFuncStr(fName(), err.Error())
					return err
				}
				err := errors.New(C.BotMessage.IssueLess)
				return err
			}
			if update.Message.Text == C.KbBtnText.CreateIssue {
				if SU.isTicketCreationPrepare() == true {
					LogFuncStr(fName(), "IssueCreatePrepare == true 1")
					SU.Messages.CreationMessage = true
					SU.updateSavedUsers()
					msg := MessageHTML(update, C.BotMessage.IssueAbout, "", false)
					_, err = api.Send(ctx, msg)
					if err != nil {
						LogFuncStr(fName(), err.Error())
						return err
					}
					return err
				}
			}
			SU.IssueID, err = SU.CreateIssue(update.Message.Text)
			SU.IssueCreatedCount = (SU.IssueCreatedCount + 1)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				msg := MessageHTML(update, C.BotMessage.IssueError, "", false)
				_, err = api.Send(ctx, msg)
				if err != nil {
					LogFuncStr(fName(), err.Error())
					return err
				}
				return err
			}
			msg := MessageHTML(update, fmt.Sprintf("%s %s", C.BotMessage.IssueCreate, SU.IssueID), "", false)
			_, err = api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}

			SU.Messages.FinishMessage = true
			SU.IssueCreated = true
			SU.IssueCreateTime = time.Now()
			SU.updateSavedUsers()
			LogFuncStr(fName(), "################# ISSUE MESSAGE #################")
			LogFuncStr(fName(), SU.PhoneNumber)
			LogFuncStr(fName(), SU.UserName)
			LogFuncStr(fName(), update.Message.Text)
			LogFuncStr(fName(), "################# ISSUE MESSAGE #################")
		}
		if update.Message.Text == C.KbBtnText.CreateIssue {
			if SU.isTicketCreationPrepare() == true {
				LogFuncStr(fName(), "IssueCreatePrepare == true 2")
				SU.Messages.CreationMessage = true
				SU.updateSavedUsers()
				msg := MessageHTML(update, C.BotMessage.IssueAbout, "", false)
				_, err = api.Send(ctx, msg)
				if err != nil {
					LogFuncStr(fName(), err.Error())
					return err
				}
			}
		}
		LogFuncStr(fName(), "TicketCreationTimeout2")
		timeout, err := SU.isTicketCreationTimeout()
		if err != nil {
			LogFuncStr(fName(), err.Error())
			msg := MessageHTML(update, fmt.Sprintf("%s %s", C.BotMessage.IssueTimeout, timeout), "", false)
			_, err = api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}
			return err //!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		} else {
			if SU.isTicketCreationPrepare() == true {
				LogFuncStr(fName(), "IssueCreatePrepare == true 3")
				err = errors.New("TicketCreationPrepare true")
				return err
			}
			user := fmt.Sprintf("%s %s", SU.FirstName, SU.MiddleName)
			msg := MessageHTML(update, fmt.Sprintf(C.BotMessage.KnownUser, user), C.KbBtnText.CreateIssue, false)
			_, err = api.Send(ctx, msg)
			if err != nil {
				LogFuncStr(fName(), err.Error())
				return err
			}
			SU.Messages.InitMessage = C.KbBtnText.CreateIssue
			SU.IssueCreatePrepare = true
			SU.IssueCreateTime = time.Now()
			SU.updateSavedUsers()
		}
	}
	return nil
}

