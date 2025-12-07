package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cppcho/slack2md/internal/adapter/presenter"
	"github.com/cppcho/slack2md/internal/adapter/repository"
	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
	"github.com/cppcho/slack2md/internal/infrastructure/config"
	"github.com/cppcho/slack2md/internal/infrastructure/filesystem"
	"github.com/cppcho/slack2md/internal/infrastructure/logger"
	infraslack "github.com/cppcho/slack2md/internal/infrastructure/slack"
	"github.com/cppcho/slack2md/internal/usecase/dto"
	"github.com/cppcho/slack2md/internal/usecase/service"
)

func main() {
	// 1. Load configuration from environment
	envConfig, err := config.LoadFromEnv()
	if err != nil {
		fmt.Printf("✗ Configuration error: %v\n", err)
		os.Exit(1)
	}

	// 2. Validate configuration
	validator := config.NewConfigValidator()
	if err := validator.Validate(envConfig); err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}

	// 3. Create infrastructure layer
	log := logger.NewLogger(logger.INFO)
	slackClient := infraslack.NewSlackClient(envConfig.BotToken, envConfig.AppToken)
	fileWriter := filesystem.NewFileSystemWriter()
	userCache := infraslack.NewUserCache()

	// 4. Create adapter layer
	slackRepo := repository.NewSlackRepository(slackClient, log, userCache)
	fileRepo := repository.NewFileRepository(fileWriter, log)
	consolePresenter := presenter.NewConsolePresenter()

	// 5. Create use case service
	exportService := service.NewExportService(slackRepo, fileRepo, log)

	// 6. Build input DTO
	input := dto.ExportChannelsInput{
		BotToken:   envConfig.BotToken,
		AppToken:   envConfig.AppToken,
		ChannelIDs: envConfig.ChannelIDs,
		ExportPath: envConfig.ExportPath,
		DaysBack:   envConfig.DaysBack,
	}

	// Print export information
	ctx := context.Background()

	// Auto-discover channels to display them
	log.Info("Auto-discovering channels...")
	channels, err := slackRepo.FetchAllChannels(ctx)
	if err != nil {
		consolePresenter.Error(fmt.Sprintf("Channel discovery failed: %v", err))
		os.Exit(1)
	}

	// If specific channels requested, filter the display
	if len(input.ChannelIDs) > 0 {
		channelIDSet := make(map[string]struct{})
		for _, id := range input.ChannelIDs {
			channelIDSet[id] = struct{}{}
		}

		var filtered []entities.Channel
		for _, ch := range channels {
			if _, exists := channelIDSet[ch.ID]; exists {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}

	consolePresenter.PrintChannelList(channels)

	// Calculate time range for display
	timeRange, _ := valueobjects.NewTimeRangeFromDaysBack(input.DaysBack)
	consolePresenter.PrintExportInfo(
		input.DaysBack,
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
		input.ExportPath,
	)

	// 7. Execute use case
	output, err := exportService.ExportChannels(ctx, input)
	if err != nil {
		consolePresenter.Error(fmt.Sprintf("Export failed: %v", err))
		os.Exit(1)
	}

	// 8. Present results
	consolePresenter.PrintSummary(output.Summary)

	if output.Summary.FailureCount > 0 {
		consolePresenter.Error("Some channels failed to export")
		os.Exit(1)
	}

	consolePresenter.Success("All channels exported successfully!")
}
