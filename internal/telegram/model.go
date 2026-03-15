package telegram

type Update struct {
	UpdateID	 int `json:"update_id"`
	Message 	*Message `json:"message"`
}

type Message struct {
	MessageID 	int `json:"message_id"`
	Text 		string `json:"text"`
	Chat 		Chat `json:"chat"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type SendMessageReq struct {
	ChatID 	int64 `json:"chat_id"`
	Text 	string `json:"text"`
}

func (u *Update) HasTextMessage() bool {
	return u.Message != nil && u.Message.Text != ""
}
