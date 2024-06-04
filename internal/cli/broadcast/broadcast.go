package broadcast

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/volvinbur1/call-of-telegram/internal/tg"
	"os"
	"strconv"
)

func NewCmd() *cobra.Command {
	broadcastCmd := &cobra.Command{
		Use:   "broadcast",
		Short: "Broadcasts message to community users",
		Long:  "Broadcast command sends certain message to users of specified community (group/channel)",
		Run:   Execute,
	}

	broadcastCmd.Flags().String("group-name", "", "Name of the group to broadcast")
	broadcastCmd.Flags().String("msg-file", "", "Path to file with target message to broadcast")

	return broadcastCmd
}

func Execute(cmd *cobra.Command, _ []string) {
	groupName, err := cmd.Flags().GetString("group-name")
	if err != nil {
		panic(fmt.Sprintf("group-name should be set: %s", err))
	}

	msgFile, err := cmd.Flags().GetString("msg-file")
	if err != nil {
		panic(fmt.Sprintf("msg-file should be set: %s", err))
	}

	apiIdStr := os.Getenv("API_ID")
	apiId, err := strconv.ParseInt(apiIdStr, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("API_ID conversion failed: %s", err))
	}

	tgApp, err := tg.NewApp(int32(apiId), os.Getenv("API_HASH"))
	if err != nil {
		panic(fmt.Sprintf("tg App creation failed: %s", err))
	}

	err = tgApp.SendMessageToGroupUsers(groupName, msgFile)
	if err != nil {
		panic(fmt.Sprintf("SendMessageToGroupUsers failed: %s", err))
	}
}
