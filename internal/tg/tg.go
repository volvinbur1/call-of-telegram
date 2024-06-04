package tg

import (
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
		NewVerbosityLevel: 1,
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
}

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
			fmt.Println("target message send failed:", err)
		} else {
			fmt.Println("Message send to", userId)
		}
		time.Sleep(5 * time.Second)
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

		membersId = append(membersId, user.UserId)
	}
	return membersId, nil
}

func (a *App) getSupergroupMembers(chatId int64) ([]*tglib.ChatMember, error) {
	offset := int32(0)
	var chatMembers []*tglib.ChatMember
	for {
		members, err := a.tgClient.GetSupergroupMembers(&tglib.GetSupergroupMembersRequest{
			SupergroupId: chatId,
			Offset:       offset,
			Limit:        200,
		})
		if err != nil {
			return nil, fmt.Errorf("supergroup members get failed: %s", err)
		}

		chatMembers = append(chatMembers, members.Members...)

		if len(members.Members) == 200 {
			offset += 200
		} else {
			return chatMembers, nil
		}
	}
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

	_, err = a.tgClient.SendMessage(&tglib.SendMessageRequest{
		ChatId: newChat.Id,
		InputMessageContent: &tglib.InputMessageText{
			Text: &tglib.FormattedText{Text: string(message)},
		},
	})
	if err != nil {
		return fmt.Errorf("send message to user %d in chat %d failed: %s", userId, newChat.Id, err)
	}

	return nil
}
