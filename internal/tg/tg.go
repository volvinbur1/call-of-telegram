package tg

import (
	"errors"
	"fmt"
	tglib "github.com/zelenin/go-tdlib/client"
	"path/filepath"
	"time"
)

type chatType int

const (
	superGroupChatType chatType = iota
	basicGroupChatType
)

func (s chatType) String() string {
	switch s {
	case basicGroupChatType:
		return "BasicGroup"
	case superGroupChatType:
		return "SuperGroup"
	default:
		return fmt.Sprintf("UnknownChatType(%d)", s)
	}
}

type ChatInfo struct {
	id  int64
	typ chatType
}

type App struct {
	tgClient *tglib.Client
	meUser   *tglib.User
}

func NewApp(apiId int32, apiHash string) (*App, error) {
	app := &App{}
	authorizer := app.createAuthorizer(apiId, apiHash)

	_, err := tglib.SetLogVerbosityLevel(&tglib.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 2,
	})
	if err != nil {
		return nil, fmt.Errorf("SetLogVerbosityLevel failed: %s", err)
	}

	app.tgClient, err = tglib.NewClient(authorizer)
	if err != nil {
		return nil, fmt.Errorf("new telegram client creation failed: %s", err)
	}

	app.meUser, err = app.tgClient.GetMe()
	if err != nil {
		return nil, fmt.Errorf("GetMe failed: %s", err)
	}
	return app, nil
	//messages, err := app.tgClient.GetChatHistory(&tglib.GetChatHistoryRequest{
	//	ChatId:    7191897364,
	//	Limit:     100,
	//	OnlyLocal: false,
	//})
	//if err != nil {
	//	return nil, fmt.Errorf("GetChatHistory failed: %s", err)
	//}
	//for _, msg := range messages.Messages {
	//	fmt.Println(msg.Id, msg.Content.(*tglib.MessageText).Text.Text)
	//	failedMsg, isOkay := msg.SendingState.(*tglib.MessageSendingStateFailed)
	//	if isOkay {
	//		fmt.Println("Error message sending failed:", failedMsg.Error)
	//	}
	//}
	//user, err := app.tgClient.GetUser(&tglib.GetUserRequest{
	//	UserId: 7191897364,
	//})
	//if err != nil {
	//	return nil, fmt.Errorf("GetUser failed: %s", err)
	//}
	//fmt.Println(user.FirstName, user.Usernames.ActiveUsernames)
	//app.tgClient.GetListener()
	//newMsg, err := app.tgClient.SendMessage(&tglib.SendMessageRequest{
	//	ChatId: 1273073098,
	//	InputMessageContent: &tglib.InputMessageText{
	//		Text: &tglib.FormattedText{Text: "Hello" + user.FirstName},
	//	},
	//})
	//if err != nil {
	//	return nil, fmt.Errorf("SendMessage failed: %s", err)
	//}
	//fmt.Println(newMsg.SendingState)
	//return nil, app.getAllChats()
}

//func (a *App) getAllChats() error {
//	chats, err := a.tgClient.GetChats(&tglib.GetChatsRequest{ChatList: &tglib.ChatListMain{},
//		Limit: 40})
//	if err != nil {
//		return fmt.Errorf("GetChats failed: %s", err)
//	}
//
//	for _, chat := range chats.ChatIds {
//		chatInfo, err := a.tgClient.GetChat(&tglib.GetChatRequest{ChatId: chat})
//		if err != nil {
//			return fmt.Errorf("GetChats failed: %s", err)
//		}
//		fmt.Println("Chat:", chatInfo.Title)
//	}
//
//	return nil
//}

func (a *App) createAuthorizer(apiId int32, apiHash string) tglib.AuthorizationStateHandler {
	authorizer := tglib.ClientAuthorizer()
	go tglib.CliInteractor(authorizer)

	authorizer.TdlibParameters <- &tglib.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseMessageDatabase:  true,
		UseSecretChats:      false,
		ApiId:               apiId,
		ApiHash:             apiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}

	return authorizer
}

func (a *App) SendMessageToGroupUsers(groupName string, message []byte) error {
	chatInfo, err := a.getChatId(groupName)
	if err != nil {
		return fmt.Errorf("get chat id of '%s' failed: %s", groupName, err)
	}

	fmt.Printf("ChatId of group `@%s` found: %d(%s)\n", groupName, chatInfo.id, chatInfo.typ)

	chatMembersIds, err := a.getChatMembersIds(chatInfo)
	if err != nil {
		return fmt.Errorf("chat members ids get failed for '%s(%d)' failed: %s", groupName, chatInfo.id, err)
	}

	fmt.Printf("Total count of users in chat `@%s`(%d): %d\n", groupName, chatInfo.id, len(chatMembersIds))

	for _, userId := range chatMembersIds {
		err = a.sendMessageToUser(userId, message)
		if err != nil {
			return errors.New("target message send failed: " + err.Error())
		}
		fmt.Println("Message send to", userId)
		time.Sleep(time.Minute)
	}
	return nil
}

func (a *App) getChatId(groupName string) (ChatInfo, error) {
	chat, err := a.tgClient.SearchPublicChat(&tglib.SearchPublicChatRequest{
		Username: groupName,
	})
	if err != nil {
		return ChatInfo{}, fmt.Errorf("search public chat by query '%s' failed: %s", groupName, err.Error())
	}

	return chatIdByType(chat)
}

func chatIdByType(chat *tglib.Chat) (ChatInfo, error) {
	if superGroup, isOkay := chat.Type.(*tglib.ChatTypeSupergroup); isOkay {
		fmt.Printf("Target chat is Supergroup with id %d\n", superGroup.SupergroupId)
		return ChatInfo{
			id:  superGroup.SupergroupId,
			typ: superGroupChatType,
		}, nil
	} else if basicGroup, isOkay := chat.Type.(*tglib.ChatTypeBasicGroup); isOkay {
		fmt.Printf("Target chat is BasicGroup with id %d\n", basicGroup.BasicGroupId)
		return ChatInfo{
			id:  basicGroup.BasicGroupId,
			typ: basicGroupChatType,
		}, nil
	} else {
		return ChatInfo{}, fmt.Errorf("unsupported chat type")
	}
}

func (a *App) getChatMembersIds(chatInfo ChatInfo) ([]int64, error) {
	var chatMembers []*tglib.ChatMember
	var err error
	if chatInfo.typ == superGroupChatType {
		chatMembers, err = a.getSupergroupMembers(chatInfo.id)
	} else {
		chatMembers, err = a.getBasicGroupMembers(chatInfo.id)
	}
	if err != nil {
		return nil, fmt.Errorf("chat members retrieve failed: %s", err)
	}

	var membersId []int64
	for _, member := range chatMembers {
		user, isOkay := member.MemberId.(*tglib.MessageSenderUser)
		if !isOkay {
			continue
		}
		if user.UserId == a.meUser.Id {
			continue
		}
		if !a.isUserRegular(user.UserId) {
			continue
		}

		membersId = append(membersId, user.UserId)
	}
	return membersId, nil
}

func (a *App) isUserRegular(userId int64) bool {
	userInfo, err := a.tgClient.GetUser(&tglib.GetUserRequest{UserId: userId})
	if err != nil {
		fmt.Printf("get user info by id %d failed: %s\n", userId, err)
		return false
	}

	_, isOkay := userInfo.Type.(*tglib.UserTypeRegular)
	if !isOkay {
		return false
	}

	canSendResult, err := a.tgClient.CanSendMessageToUser(&tglib.CanSendMessageToUserRequest{
		UserId:    userId,
		OnlyLocal: false,
	})
	if err != nil {
		fmt.Printf("CanSendMessageToUser %d failed: %s\n", userId, err)
	}
	_, isOkay = canSendResult.(*tglib.CanSendMessageToUserResultOk)
	return isOkay
}

func (a *App) getSupergroupMembers(chatId int64) ([]*tglib.ChatMember, error) {
	var chatMembers []*tglib.ChatMember
	members, err := a.tgClient.GetSupergroupMembers(&tglib.GetSupergroupMembersRequest{
		Filter:       &tglib.SupergroupMembersFilterMention{},
		SupergroupId: chatId,
		Limit:        200,
	})
	if err != nil {
		return nil, fmt.Errorf("supergroup members get failed: %s", err)
	}

	chatMembers = append(chatMembers, members.Members...)

	return chatMembers, nil
}

func (a *App) getBasicGroupMembers(chatId int64) ([]*tglib.ChatMember, error) {
	basicChatInfo, err := a.tgClient.GetBasicGroupFullInfo(&tglib.GetBasicGroupFullInfoRequest{
		BasicGroupId: chatId,
	})
	if err != nil {
		return nil, fmt.Errorf("basic group full info get failed: %s", err)
	}
	return basicChatInfo.Members, nil
}

func (a *App) sendMessageToUser(userId int64, message []byte) error {
	newChat, err := a.tgClient.CreatePrivateChat(&tglib.CreatePrivateChatRequest{
		UserId: userId,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("creation of private chat failed: %s", err)
	}

	msg, err := a.tgClient.SendMessage(&tglib.SendMessageRequest{
		ChatId: newChat.Id,
		InputMessageContent: &tglib.InputMessageText{
			Text: &tglib.FormattedText{Text: string(message)},
		},
	})
	if err != nil {
		return fmt.Errorf("send message to user %d failed: %s", userId, err)
	}

	return a.handleMessageState(msg)
}

func (a *App) handleMessageState(message *tglib.Message) error {
	switch message.SendingState.MessageSendingStateType() {
	case tglib.TypeMessageSendingStateFailed:
		failState, isOkay := message.SendingState.(*tglib.MessageSendingStateFailed)
		if !isOkay {
			return fmt.Errorf("messageSendingStateFailed state is not convertible to *tglib.MessageSendingStateFailed")
		}
		return fmt.Errorf("message %d to user %d in FAILED state. ErrCode: %d ErrMessage %s", message.Id, message.ChatId, failState.Error.Code, failState.Error.Message)
	case tglib.TypeMessageSendingStatePending:
		fmt.Printf("message %d to user %d in PEDNING state\n", message.Id, message.ChatId)
		return a.listenToNewEventsAfterSendingMessage()
	default:
		return fmt.Errorf("message %d to user %d has unsupported state: %s\n", message.Id, message.ChatId, message.SendingState.MessageSendingStateType())
	}
}

func (a *App) listenToNewEventsAfterSendingMessage() error {
	for update := range a.tgClient.GetListener().Updates {
		fmt.Println("New EVENT:", update.GetType())
		switch update.GetType() {
		case tglib.TypeUpdateMessageSendFailed:
			upd, isOkay := update.(*tglib.UpdateMessageSendFailed)
			if !isOkay {
				return fmt.Errorf("updateMessageSendFailed event is not convertible to *tglib.UpdateMessageSendFailed")
			}
			return fmt.Errorf("message %d sent to user %d failed. ErrCode: %d ErrMsg: %s", upd.OldMessageId, upd.Message.ChatId, upd.Error.Code, upd.Error.Message)
		case tglib.TypeUpdateMessageSendSucceeded:
			upd, isOkay := update.(*tglib.UpdateMessageSendSucceeded)
			if !isOkay {
				return fmt.Errorf("updateMessageSendSucceeded event is not convertible to *tglib.UpdateMessageSendSucceeded")
			}
			fmt.Printf("Message %d sent to user %d SUCCESSFULLY\n", upd.Message.Id, upd.Message.ChatId)
			return nil
		case tglib.TypeUpdateDeleteMessages:
			upd, isOkay := update.(*tglib.UpdateDeleteMessages)
			if !isOkay {
				return fmt.Errorf("updateDeleteMessages event is not convertible to *tglib.UpdateDeleteMessages")
			}
			fmt.Printf("Messages %v from chat %d were deleted\n", upd.MessageIds, upd.ChatId)
			return nil
		}
	}
	return nil
}
