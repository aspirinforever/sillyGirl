package tg

import (
	"net/http"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/cdle/sillyGirl/im"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Sender struct {
	Message *tb.Message
	matches [][]string
}

var b *tb.Bot
var cfg *im.Config
var Handler = func(message *tb.Message) {

}

func RunBot(conf *im.Config) {
	cfg = conf
	var err error
	b, err = tb.NewBot(tb.Settings{
		Token:  cfg.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		logs.Warn("监听tgbot失败：%v", err)
		return
	}
	b.Handle(tb.OnText, Handler)
	logs.Info("监听tgbot")
	b.Start()
}

func (sender *Sender) GetContent() string {
	return sender.Message.Text
}

func (sender *Sender) GetUserID() int {
	return sender.Message.Sender.ID
}

func (sender *Sender) GetChatID() int {
	return int(sender.Message.Chat.ID)
}

func (sender *Sender) GetImType() string {
	return "tg"
}

func (sender *Sender) GetMessageID() int {
	return sender.Message.ID
}

func (sender *Sender) GetUsername() string {
	return sender.Message.Sender.Username
}

func (sender *Sender) IsReply() bool {
	return sender.Message.ReplyTo != nil
}

func (sender *Sender) GetReplySenderUserID() int {
	if !sender.IsReply() {
		return 0
	}
	return sender.Message.ReplyTo.ID
}

func (sender *Sender) GetRawMessage() interface{} {
	return sender.Message
}

func (sender *Sender) SetMatch(ss []string) {
	sender.matches = [][]string{ss}
}
func (sender *Sender) SetAllMatch(ss [][]string) {
	sender.matches = ss
}

func (sender *Sender) GetMatch() []string {
	return sender.matches[0]
}

func (sender *Sender) GetAllMatch() [][]string {
	return sender.matches
}

func (sender *Sender) Get(index ...int) string {

	i := 0
	if len(index) != 0 {
		i = index[0]
	}
	if len(sender.matches) == 0 {
		return ""
	}
	if len(sender.matches[0]) < i+1 {
		return ""
	}
	return sender.matches[0][i]
}

func (sender *Sender) IsAdmin() bool {
	for _, id := range cfg.Masters {
		if id == sender.Message.Sender.ID {
			return true
		}
	}
	return false
}

func (sender *Sender) IsMedia() bool {
	return false
}

func (sender *Sender) Reply(rt interface{}) error {
	var r tb.Recipient
	if !sender.Message.FromGroup() {
		r = sender.Message.Sender
	} else {
		r = sender.Message.Chat
	}
	switch rt.(type) {
	case error:
		b.Send(r, rt.(error).Error, &tb.SendOptions{ReplyTo: sender.Message})
	case []byte:
		b.Send(r, string(rt.([]byte)), &tb.SendOptions{ReplyTo: sender.Message})
	case string:
		b.Send(r, rt.(string), &tb.SendOptions{ReplyTo: sender.Message})
	case *http.Response:
		b.SendAlbum(r, tb.Album{&tb.Photo{File: tb.FromReader(rt.(*http.Response).Body)}}, &tb.SendOptions{ReplyTo: sender.Message})
	}
	return nil
}
