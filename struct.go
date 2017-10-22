package main

import (
	"time"
	"github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
)

type Config struct {
	DB DB
	Log Log
	Serv Serv
	StartMsg StartMsg
	KbBtnText KbBtnText
	BotMessage BotMessage
	JiraIssue JiraIssue
}

type DB struct {
	Host	string
	Port	string
	Name	string
	User	string
	Pass	string
	SSL	string
	Type	string
}

type Log struct {
	Dir	string
	File	string
}

type Serv struct {
	Token	string
	Debug	bool
}

type StartMsg struct {
	Contact		string
	Command		string
}

type KbBtnText struct {
	Auth		string
	CreateIssue	string
}

type BotMessage struct {
	NeedAuth	string
	KnownUser	string
	ContactNotOwn	string
	IssueAbout	string
	IssueTimeout	string
	IssueLimit	string
	IssueError	string
	IssueCreate	string
	IssueLess	string
}

type JiraIssue struct {
	Project			string
	TypeID			string
	Summary			string
	UrlUser			string
	UrlPass			string
	UrlScheme		string
	UrlHost			string
	UrlPathUserSearch	string
	UrlPathCreateIssue	string
	Limit			int
	Timeout			float64
	DescLen			int
}

type TelegramIssueUsers struct {
	rec		structable.Recorder
	builder 	squirrel.StatementBuilderType

	Id		int		`stbl:"id,PRIMARY_KEY,SERIAL"`
	UserId		int64		`stbl:"user_id"`
	PhoneNumber	string		`stbl:"phone_number"`
	FirstName	string		`stbl:"first_name"`
	LastName	string		`stbl:"last_name"`
	UserName	string		`stbl:"user_name"`
	MiddleName	string		`stbl:"middle_name"`
	Status		int		`stbl:"status"`
	UpdatedAt	int64		`stbl:"updated_at"`
}

type conventionalMarshaller struct {
	Value interface{}
}

type SavedUser struct {
	TgID			int64
	PhoneNumber		string
	UserName		string
	FirstName		string
	MiddleName		string
	LastName		string
	Location		string
	IssueCreateTime		time.Time
	IssueCreatePrepare	bool
	IssueCreated		bool
	IssueCreatedCount	int
	Messages		UserMessages
	IssueID			string
}

type UserMessages struct {
	InitMessage		string
	CreationMessage		bool
	FinishMessage		bool
}

type IssueResponse struct {
	Id	string
	Key	string
	Self	string
}

type IssueResponseJ struct {
	Id	string	`json:"id"`
	Key	string	`json:"key"`
	Self	string	`json:"self"`
}

type IssueCreate struct {
	Fields		IssueCreateProject
}

type IssueCreateProject struct {
	Project			IssueCreateKey
	Summary			string
	Description		string
	Issuetype		IssueCreateId
	Reporter		IssueReporter
}

type IssueCreateKey struct {
	Key	string
}

type IssueCreateId struct {
	Id	string
}

type  IssueReporter struct {
	Name	string
}

type JiraUserJ struct {
	Users	JiraUsersSub
}

type JiraUsersSub struct {
	Users	[]JiraUsersSub2

}
type JiraUsersSub2 struct {
	Name	string	`json:"name"`
}
