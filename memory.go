package deepbot

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/wdvxdr1123/ZeroBot"
)

type msgType struct {
	UserID  uint64 `json:"user_id"`
	GroupID uint64 `json:"group_id"`
	SubType string `json:"sub_type"`
	Time    uint64 `json:"time"`
	Sender  struct {
		UserID   uint64 `json:"user_id"`
		Nickname string `json:"nickname"`
		Card     string `json:"card,omitempty"`
		Role     string `json:"role,omitempty"`
		Sex      string `json:"sex,omitempty"`
		Age      uint64 `json:"age,omitempty"`
	} `json:"sender"`
	Message []struct {
		Type string `json:"type"`
		Data struct {
			Text    string      `json:"text,omitempty"`
			Name    string      `json:"name,omitempty"`
			QQ      string      `json:"qq"`
			ID      string      `json:"id"`
			File    string      `json:"file,omitempty"`
			Summary string      `json:"summary,omitempty"`
			Data    string      `json:"data,omitempty"`
			Content interface{} `json:"content,omitempty"`
		} `json:"data"`
	} `json:"message"`
	MessageID       uint64 `json:"message_id"`
	MessageSeq      uint64 `json:"message_seq"`
	MessageFormat   string `json:"message_format"`
	MessageType     string `json:"message_type"`
	MessageSentType string `json:"message_sent_type"`
	PostType        string `json:"post_type"`
	RealID          uint64 `json:"real_id"`
	SelfID          uint64 `json:"self_id"`
	RawMessage      string `json:"raw_message"`
	Font            uint64 `json:"font"`
}

func (bot *DeepBot) generateSTM(ctx *zero.Ctx) {
	params := make(zero.Params)
	params["group_id"] = bot.config.GroupID[0]
	params["message_seq"] = 0
	params["count"] = 100
	params["reverseOrder"] = false
	resp := ctx.CallAction("get_group_msg_history", params)
	if resp.Status != "ok" {
		return
	}

	var messages []msgType
	raw := resp.Data.Get("messages").Raw
	err := json.NewDecoder(strings.NewReader(raw)).Decode(&messages)
	if err != nil {
		log.Println("failed to read group history message:", err)
		return
	}
	for _, msg := range messages {
		fmt.Println(msg.Time, msg.Sender.Nickname, msg.Sender.Card, msg.Sender.UserID)
		for _, m := range msg.Message {
			switch m.Type {
			case "text":
				fmt.Println("text:", m.Data.Text)
			case "at":
				fmt.Println("at:", m.Data.QQ, m.Data.Name)
			case "reply":
				fmt.Println("reply:", m.Data.ID)
			default:
				fmt.Println(m.Type)
			}
		}
	}
}
