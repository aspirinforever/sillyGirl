package core

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/cdle/sillyGirl/im"
	"github.com/cdle/sillyGirl/im/tg"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Function struct {
	Rules   []string
	FindAll bool
	Admin   bool
	Handle  func(s im.Sender) bool
}

var pname = regexp.MustCompile(`/([^/\s]+)`).FindStringSubmatch(os.Args[0])[1]

var functions = []Function{
	{
		Rules: []string{"^傻妞 (.*)$", "^傻妞$"},
		Handle: func(s im.Sender) bool {
			m := s.Get()
			if m != "" {
				s.Reply(fmt.Sprintf("哎呀，傻妞不懂%s是什么意思啦。", m))
			} else {
				s.Reply("请说，我在。")
			}
			return true
		},
	},
	{
		Rules: []string{"^升级$"},
		Admin: true,
		Handle: func(s im.Sender) bool {
			s.Reply("傻妞开始拉取代码。")
			rtn, err := exec.Command("sh", "-c", "cd "+ExecPath+" && git stash && git pull").Output()
			if err != nil {
				s.Reply("傻妞拉取代失败：" + err.Error() + "。")
				return true
			}
			t := string(rtn)
			if !strings.Contains(t, "changed") {
				if strings.Contains(t, "Already") || strings.Contains(t, "已经是最新") {
					s.Reply("傻妞已是最新版啦。")
					return true
				} else {
					s.Reply("傻妞拉取代失败：" + t + "。")
					return true
				}
			} else {
				s.Reply("傻妞拉取代码成功。")
			}
			s.Reply("傻妞正在编译程序。")
			rtn, err = exec.Command("sh", "-c", "cd "+ExecPath+" && go build -o "+pname).Output()
			if err != nil {
				s.Reply("傻妞编译失败：" + err.Error())
				return true
			} else {
				s.Reply("傻妞编译成功。")
			}
			s.Reply("傻妞重启程序。")
			Daemon()
			return true
		},
	},
	{
		Rules: []string{"^重启$"},
		Admin: true,
		Handle: func(s im.Sender) bool {
			s.Reply("傻妞重启程序。")
			Daemon()
			return true
		},
	},
}

var Senders chan im.Sender

func initToHandleMessage() {
	if len(Config.Im) == 0 {
		logs.Warn("未配置置通讯工具")
	}
	for _, im := range Config.Im {
		switch im.Type {
		case "tg":
			tg.Handler = func(message *tb.Message) {
				Senders <- &tg.Sender{
					Message: message,
				}
			}
			go tg.RunBot(&im)
		case "qq":

		}
	}
	Senders = make(chan im.Sender)
	go func() {
		for {
			go handleMessage(<-Senders)
		}
	}()
}

func AddCommand(cmd *Function) {
	functions = append(functions, *cmd)
}

func handleMessage(sender im.Sender) {
	for _, function := range functions {
		for _, rule := range function.Rules {
			var matched bool
			if function.FindAll {
				if res := regexp.MustCompile(rule).FindAllStringSubmatch(sender.GetContent(), -1); len(res) > 0 {
					tmp := [][]string{}
					for i := range res {
						tmp = append(tmp, res[i][1:])
					}
					sender.SetAllMatch(tmp)
					matched = true
				}
			} else {
				if res := regexp.MustCompile(rule).FindStringSubmatch(sender.GetContent()); len(res) > 0 {
					sender.SetMatch(res[1:])
					matched = true
				}
			}
			if matched && function.Handle(sender) {
				return
			}
		}
	}
}
