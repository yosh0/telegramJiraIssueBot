package main

import (
	"os"
	"fmt"
	"log"
	"path"
	"runtime"
	"strconv"
)

func defaultLog() string {
	_, file, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(file), C.Log.File)
}

func LogFuncStr(fName, text string) {
	f, err := os.OpenFile(fmt.Sprintf("%s%s", C.Log.Dir, C.Log.File), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		f, err = os.OpenFile(defaultLog(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	}
	log.SetOutput(f)
	log.Println(fName, text)
	f.Close()
}

func LogSavedUsers(fName string, user SavedUser) {
	f, err := os.OpenFile(fmt.Sprintf("%s%s", C.Log.Dir, C.Log.File), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		f, err = os.OpenFile(defaultLog(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	}
	log.SetOutput(f)
	log.Println(fName, "{\n" +
		"	TgID: "+strconv.FormatInt(user.TgID, 10)+"\n"+
		"	PhoneNumber: "+user.PhoneNumber+"\n"+
		"	UserName: "+user.UserName+"\n"+
		"	FirstName: "+user.FirstName+"\n"+
		"	MiddleName: "+user.MiddleName+"\n"+
		"	LastName: "+user.LastName+"\n"+
		"	Location: "+user.Location+"\n"+
		"	IssueCreateTime: "+user.IssueCreateTime.String()+"\n"+
		"	IssueCreatePrepare: "+strconv.FormatBool(user.IssueCreatePrepare)+"\n"+
		"	IssueCreated: "+strconv.FormatBool(user.IssueCreated)+"\n"+
		"	IssueCreatedCount: "+strconv.Itoa(user.IssueCreatedCount)+"\n"+
		"	IssueID: "+user.IssueID+"\n"+
		"	Messages{\n" +
		"		InitMessage: "+user.Messages.InitMessage+"\n"+
		"		CreationMessage: "+strconv.FormatBool(user.Messages.CreationMessage)+"\n"+
		"		FinishMessage: "+strconv.FormatBool(user.Messages.FinishMessage)+"\n"+
		"	}\n"+
		"}",
	)
	f.Close()
}
