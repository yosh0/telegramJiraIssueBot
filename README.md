## USED STRUCTS
- [SQL](#sql)
- [DB Struct](#db-struct)
- [Saved Users](#saved-users)
- [Issue Create](#issue-create)
- [Issue Response](#issue-response)

## INFO
- [Library Info](#library-info)

## GitHub Links
- [Telegram](https://github.com/bot-api/telegram)

##### Package contains:
1. Client for telegram bot api.
2. Bot with:
    1. Middleware support
        1. Command middleware to handle commands.
        2. Recover middleware to recover on panics.
    2. Webhook support

#### `SQL`
```
CREATE SEQUENCE telegram_issue_users_id_seq
    INCREMENT 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    START 1
    CACHE 1;
ALTER TABLE telegram_issue_users_id_seq OWNER TO user;

CREATE TABLE telegram_issue_users
(
    id serial NOT NULL,
    user_id bigint,
    phone_number character varying(32),
    first_name character varying(32),
    last_name character varying(32),
    middle_name character varying(32),
    user_name character varying(32),
    updated_at integer,
    status smallint
)
WITH (
    OIDS=FALSE
);
ALTER TABLE telegram_issue_users OWNER TO user;
```

#### `DB Struct`
```
type TelegramIssueUsers struct {
    rec         structable.Recorder
    builder     squirrel.StatementBuilderType
    
    Id          int     `stbl:"id,PRIMARY_KEY,SERIAL"`
    UserId      int64   `stbl:"user_id"`
    PhoneNumber string  `stbl:"phone_number"`
    FirstName   string  `stbl:"first_name"`
    LastName    string  `stbl:"last_name"`
    UserName    string  `stbl:"user_name"`
    MiddleName  string  `stbl:"middle_name"`
    Status      int     `stbl:"status"`
    UpdatedAt   int64   `stbl:"updated_at"`
}
```

#### `Saved Users`
```
SavedUser{
    TgID: id,
    PhoneNumber: phoneNumber,
    UserName: userName,
    FirstName: firstName,
    MiddleName: middleName,
    LastName: lastName,
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
    IssueID: PROJECT-0
}
```

#### `Issue Create`
```
IssueCreate{
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
```

#### `Issue Response`
```
IssueResponse{
    Id: I.Id,
    Key: I.Key,
    Self: I.Self,
}
```

#### `Info`
```
    Используется альтернативная NewKeyboardButton с параметром `request_contact`
```

```go
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
```
#### `Library Info`

##### Supported API methods:
- [x] getMe
- [x] sendMessage
- [x] forwardMessage
- [x] sendPhoto
- [x] sendAudio
- [x] sendDocument
- [x] sendSticker
- [x] sendVideo
- [x] sendVoice
- [x] sendLocation
- [x] sendChatAction
- [x] getUserProfilePhotos
- [x] getUpdates
- [x] setWebhook
- [x] getFile
- [x] answerInlineQuery inline bots

#####  Supported API v2 methods:
- [x] sendVenue
- [x] sendContact
- [x] editMessageText
- [x] editMessageCaption
- [x] editMessageReplyMarkup
- [x] kickChatMember
- [x] unbanChatMember
- [x] answerCallbackQuery
- [x] getChat
- [x] getChatMember
- [x] getChatMembersCount
- [x] getChatAdministrators
- [x] leaveChat

##### Supported Inline modes
- [x] InlineQueryResultArticle
- [x] InlineQueryResultAudio
- [x] InlineQueryResultContact
- [x] InlineQueryResultDocument
- [x] InlineQueryResultGif
- [x] InlineQueryResultLocation
- [x] InlineQueryResultMpeg4Gif
- [x] InlineQueryResultPhoto
- [x] InlineQueryResultVenue
- [x] InlineQueryResultVideo
- [x] InlineQueryResultVoice
- [ ] InlineQueryResultCachedAudio
- [ ] InlineQueryResultCachedDocument
- [ ] InlineQueryResultCachedGif
- [ ] InlineQueryResultCachedMpeg4Gif
- [ ] InlineQueryResultCachedPhoto
- [ ] InlineQueryResultCachedSticker
- [ ] InlineQueryResultCachedVideo
- [ ] InlineQueryResultCachedVoice
- [ ] InputTextMessageContent
- [ ] InputLocationMessageContent
