package broadcast

import (
	"bufio"
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

	broadcastCmd.Flags().String("group-list", "", "path to file with groups to broadcast")
	broadcastCmd.Flags().String("group-name", "", "group to broadcast")
	broadcastCmd.Flags().String("msg-file", "", "Path to file with target message to broadcast")

	return broadcastCmd
}

func Execute(cmd *cobra.Command, _ []string) {
	groupListPath, _ := cmd.Flags().GetString("group-list")
	groupName, _ := cmd.Flags().GetString("group-name")

	message := getTargetMessage(cmd)

	apiIdStr := os.Getenv("API_ID")
	apiId, err := strconv.ParseInt(apiIdStr, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("API_ID conversion failed: %s", err))
	}

	tgApp, err := tg.NewApp(int32(apiId), os.Getenv("API_HASH"))
	if err != nil {
		panic(fmt.Sprintf("tg App creation failed: %s", err))
	}

	if groupName != "" {
		err = tgApp.SendMessageToGroupUsers(groupName, message)
		if err != nil {
			panic(fmt.Sprintf("SendMessageToGroupUsers failed: %s", err))
		}
	} else if groupListPath != "" {
		err = readGroupList(groupListPath, tgApp, message)
		if err != nil {
			panic(fmt.Sprintf("ReadGroupList failed: %s", err))
		}
	} else {
		panic(fmt.Sprintf("either group-list or group-name should be set"))
	}
}

func getTargetMessage(cmd *cobra.Command) []byte {
	msgFile, err := cmd.Flags().GetString("msg-file")
	if err != nil {
		panic(fmt.Sprintf("msg-file should be set: %s", err))
	}

	data, err := os.ReadFile(msgFile)
	if err != nil {
		panic(fmt.Sprintf("read file %s failed: %s", msgFile, err))
	}

	return data
}

func readGroupList(groupListFile string, tgApp *tg.App, message []byte) error {
	file, err := os.Open(groupListFile)
	if err != nil {
		return fmt.Errorf("open file %s failed: %s", groupListFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		groupName := scanner.Text()
		fmt.Println("Read group:", groupName, "PROCESSING")
		err = tgApp.SendMessageToGroupUsers(groupName, message)
		if err != nil {
			fmt.Println("SendMessageToGroupUsers failed:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file %s failed: %s", groupListFile, err)
	}

	return nil
}
