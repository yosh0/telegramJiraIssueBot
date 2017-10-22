package main

import (
	"fmt"
	"time"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
)

func NewTelegramIssueUsers(db squirrel.DBProxyBeginner) *TelegramIssueUsers {
	d := new(TelegramIssueUsers)
	d.builder = squirrel.StatementBuilder.RunWith(db)
	if C.DB.Type == DB_DRIVER {
		d.builder = d.builder.PlaceholderFormat(squirrel.Dollar)
	}
	d.rec = structable.New(db, C.DB.Type).Bind(DB_TBL, d)
	return d
}

func (TIU *TelegramIssueUsers) LoadByUserId() error {
	return TIU.rec.LoadWhere("user_id = ?", TIU.UserId)
}

func (TIU *TelegramIssueUsers) LoadByPhoneNumber() error {
	return TIU.rec.LoadWhere("phone_number = ?", TIU.PhoneNumber)
}


func (TIU *TelegramIssueUsers) Update() error {
	return TIU.rec.Update()
}

func (SU SavedUser) NewIssuePrepare(text, login string) IssueCreate {
	return IssueCreate{
		Fields: IssueCreateProject{
			Project: IssueCreateKey{
				C.JiraIssue.Project,
			},
			Description: text,
			Summary: C.JiraIssue.Summary,
			Issuetype: IssueCreateId{
				Id: C.JiraIssue.TypeID,
			},
			Reporter: IssueReporter{
				Name: login,
			},
		},
	}
}

func NewIssueResponse(I IssueResponseJ) IssueResponse {
	return IssueResponse{
		Id: I.Id,
		Key: I.Key,
		Self: I.Self,
	}
}

func (c conventionalMarshaller) MarshalJSON() ([]byte, error) {
	marshalled, err := json.Marshal(c.Value)
	if err != nil {
		LogFuncStr(fName(), err.Error())
	}
	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			return bytes.ToLower(wordBarrierRegex.ReplaceAll(
				match,
				[]byte(`${1}_${2}`),
			))
		},
	)
	return converted, err
}

func (SU SavedUser) GetJiraUserLogin() (string, error) {
	var Url *url.URL
	Url, err := url.Parse(fmt.Sprintf("%s://%s:%s@%s", C.JiraIssue.UrlScheme, C.JiraIssue.UrlUser, C.JiraIssue.UrlPass, C.JiraIssue.UrlHost))
	if err != nil {
		LogFuncStr(fName(), err.Error())
		return "", err
	}
	Url.Scheme = C.JiraIssue.UrlScheme
	Url.Host = C.JiraIssue.UrlHost
	Url.Path = C.JiraIssue.UrlPathUserSearch
	params := url.Values{}
	params.Add("query", fmt.Sprintf("%s %s %s", SU.LastName, SU.FirstName, SU.MiddleName))
	Url.RawQuery = params.Encode()
	LogFuncStr(fName(), string(Url.String()))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := http.NewRequest("GET", Url.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	rq, err := client.Do(resp)
	LogFuncStr(fName(), rq.Status)
	defer rq.Body.Close()
	body, err := ioutil.ReadAll(rq.Body)
	if err != nil {
		LogFuncStr(fName(), err.Error())
	}
	LogFuncStr(fName(), string(body))
	var jiraUserJ JiraUserJ
	if err := json.Unmarshal([]byte(body), &jiraUserJ); err != nil {
		LogFuncStr(fName(), "\n### JSON ERR\n" +
			err.Error() + "\n" +
			"### JSON ERR\n")
		return "", err
	}
	for _, dd := range jiraUserJ.Users.Users {
		LogFuncStr(fName(), dd.Name)
		if len(dd.Name) > 0 {
			return dd.Name, nil
		}
	}
	LogFuncStr(fName(), "User not found by lastName and firstName")
	LogFuncStr(fName(), fmt.Sprintf("Return default user '%s'", C.JiraIssue.UrlUser))
	return C.JiraIssue.UrlUser, nil
}

func (SU SavedUser) CreateIssue(text string) (string, error) {
	var nullString string
	login, err := SU.GetJiraUserLogin()
	if err != nil {
		return nullString, err
	}
	jm, err := json.MarshalIndent(conventionalMarshaller{SU.NewIssuePrepare(text, login)}, "", " ")
	if err != nil {
		LogFuncStr(fName(), err.Error())
		return nullString, err
	}
	LogFuncStr(fName(), string(jm))

	var Url *url.URL
	Url, err = url.Parse(fmt.Sprintf("%s://%s:%s@%s", C.JiraIssue.UrlScheme, C.JiraIssue.UrlUser, C.JiraIssue.UrlPass, C.JiraIssue.UrlHost))
	if err != nil {
		LogFuncStr(fName(), err.Error())
		return nullString, err
	}
	Url.Scheme = C.JiraIssue.UrlScheme
	Url.Host = C.JiraIssue.UrlHost
	Url.Path = C.JiraIssue.UrlPathCreateIssue
	LogFuncStr(fName(), string(Url.String()))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	payload := strings.NewReader(string(jm))
	req, err := http.NewRequest("POST", Url.String(), payload)
	if err != nil {
		LogFuncStr(fName(), err.Error())
		err = errors.New("ISSUE SERVER UNREACHABLE")
		return nullString, err
	}
	req.Header.Add("content-type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		LogFuncStr(fName(), err.Error())
		err = errors.New("ISSUE SERVER REQUEST ERROR")
		return nullString, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		LogFuncStr(fName(), err.Error())
		err = errors.New("ISSUE SERVER RESPONSE ERROR")
		return nullString, err
	}
	LogFuncStr(fName(), string(body))
	var issueResponseJ IssueResponseJ
	if err := json.Unmarshal([]byte(body), &issueResponseJ); err != nil {
		LogFuncStr(fName(), "\n### JSON ERR\n" + err.Error() + "\n" + "### JSON ERR\n")
		err = errors.New("ISSUE SERVER RESPONSE PARSE ERROR")
		return nullString, err
	}
	issue := NewIssueResponse(issueResponseJ)
	return issue.Key, nil
}

func NewSavedUser(id int64, tiu *TelegramIssueUsers) SavedUser {
	SU := SavedUser{
		TgID: id,
		PhoneNumber: tiu.PhoneNumber,
		UserName: tiu.UserName,
		FirstName: tiu.FirstName,
		MiddleName: tiu.MiddleName,
		LastName: tiu.LastName,
		Location: "",
		IssueCreatePrepare: false,
		IssueCreateTime: time.Now(),
		IssueCreated: false,
		IssueCreatedCount: 0,
		Messages: UserMessages{
			InitMessage: "",
			CreationMessage: false,
			FinishMessage: false,
		},
		IssueID: "",
	}
	LogSavedUsers(fName(), SU)
	return SU
}

func (SU SavedUser) updateSavedUsers() {
	listSavedUsers[SU.TgID] = SavedUser{
		TgID: SU.TgID,
		PhoneNumber: SU.PhoneNumber,
		UserName: SU.UserName,
		FirstName: SU.FirstName,
		MiddleName: SU.MiddleName,
		LastName: SU.LastName,
		Location: "",
		IssueCreatePrepare: SU.IssueCreatePrepare,
		IssueCreateTime: SU.IssueCreateTime,
		IssueCreated: SU.IssueCreated,
		IssueCreatedCount: SU.IssueCreatedCount,
		Messages: UserMessages{
			InitMessage: SU.Messages.InitMessage,
			CreationMessage: SU.Messages.CreationMessage,
			FinishMessage: SU.Messages.FinishMessage,
		},
	}
	LogSavedUsers(fName(), SU)
}

func (SU SavedUser) clearSavedUsers() {
	if _, ok := listSavedUsers[SU.TgID]; ok {
		listSavedUsers[SU.TgID] = SavedUser{
			TgID: SU.TgID,
			PhoneNumber: SU.PhoneNumber,
			UserName: SU.UserName,
			FirstName: SU.FirstName,
			MiddleName: SU.MiddleName,
			LastName: SU.LastName,
			Location: "",
			IssueCreatePrepare: true,
			IssueCreateTime: time.Time{},
			IssueCreated: false,
			IssueCreatedCount: SU.IssueCreatedCount,
			Messages: UserMessages{
				InitMessage: "",
				CreationMessage: false,
				FinishMessage: false,
			},
			IssueID: "",
		}
		LogSavedUsers(fName(), SU)
	}
}

func (SU SavedUser) isTicketCreationPrepare() bool {
	if _, ok := listSavedUsers[SU.TgID]; ok {
		if listSavedUsers[SU.TgID].IssueCreatePrepare == true {
			LogFuncStr(fName(), "IssueCreatePrepare == true")
			return true
		}
	}
	return false
}

func (SU SavedUser) isCreationMessge() bool {
	if _, ok := listSavedUsers[SU.TgID]; ok {
		if listSavedUsers[SU.TgID].Messages.CreationMessage == true {
			return true
		}
	}
	return false
}

func (SU SavedUser) isTicketCreated() bool {
	if _, ok := listSavedUsers[SU.TgID]; ok {
		if listSavedUsers[SU.TgID].IssueCreated == true {
			return true
		}
	}
	return false

}

func (SU SavedUser) isTicketCreationTimeout() (string, error) {
	timeString := TimeString(C.JiraIssue.Timeout - time.Since(listSavedUsers[SU.TgID].IssueCreateTime).Seconds())
	LogFuncStr(fName(), fmt.Sprintf("Timeout - %s", timeString))
	LogFuncStr(fName(), time.Since(listSavedUsers[SU.TgID].IssueCreateTime).String())
	TimeOut := strconv.FormatFloat(time.Since(listSavedUsers[SU.TgID].IssueCreateTime).Seconds(), 'g', 10, 64)

	TimeOut2 := C.JiraIssue.Timeout - time.Since(listSavedUsers[SU.TgID].IssueCreateTime).Seconds()
	TimeOut3 := strconv.FormatFloat(TimeOut2, 'g', 10, 64)

	LogFuncStr(fName(), fmt.Sprintf("TimeOut %s", TimeOut))
	LogFuncStr(fName(), fmt.Sprintf("TimeOut3 %s", TimeOut3))

	if _, ok := listSavedUsers[SU.TgID]; ok {
		if time.Since(listSavedUsers[SU.TgID].IssueCreateTime).Seconds() < C.JiraIssue.Timeout {
			if listSavedUsers[SU.TgID].IssueCreated == true {
				LogFuncStr(fName(), fmt.Sprintf("Creation Timeout %s %s", listSavedUsers[SU.TgID].PhoneNumber, timeString))
				err := errors.New("Creation Timeout")
				return timeString, err
			} else {
				return "", nil
			}
			err := errors.New("Creation Timeout")
			return "", err
		} else {
			if listSavedUsers[SU.TgID].IssueCreated == true {
				SU.clearSavedUsers()
				SU.IssueCreated = false
				SU.IssueCreatePrepare = true
				SU.updateSavedUsers()
				LogFuncStr(fName(), "CLEAR USER FIELDS")
				return "", nil
			}
		}
	}
	return "", nil
}

func inSavedUsersList(id int64) (SavedUser, error) {
	var SU SavedUser
	fmt.Println(listSavedUsers)
	if _, ok := listSavedUsers[id]; ok {
		LogSavedUsers(fName(), listSavedUsers[id])
		return listSavedUsers[id], nil
	}
	LogFuncStr(fName(), fmt.Sprintf("NOT IN LIST %d", id))
	err := errors.New(fmt.Sprintf("NOT IN LIST %d", id))
	return SU, err
}
