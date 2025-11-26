package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/theQRL/qrl-beaconchain-explorer/services"
	"github.com/theQRL/qrl-beaconchain-explorer/types"
	"github.com/theQRL/qrl-beaconchain-explorer/utils"
	"github.com/theQRL/qrl-beaconchain-explorer/version"
)

var layoutTemplateFiles = []string{
	"layout.html",
	"layout/mainnavigation.html",
}

func InitPageData(w http.ResponseWriter, r *http.Request, active, path, title string, mainTemplates []string) *types.PageData {
	fullTitle := fmt.Sprintf("%v - %v - explorer.zond.theqrl.org - %v", title, utils.Config.Frontend.SiteName, time.Now().Year())

	if title == "" {
		fullTitle = fmt.Sprintf("%v - explorer.zond.theqrl.org - %v", utils.Config.Frontend.SiteName, time.Now().Year())
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fullTitle,
			Description: "explorer.zond.theqrl.org makes QRL accessible to non-technical end users",
			Path:        path,
			GATag:       utils.Config.Frontend.GATag,
			NoTrack:     false,
			Templates:   strings.Join(mainTemplates, ","),
		},

		Active:                active,
		Data:                  &types.Empty{},
		Version:               version.Version,
		Year:                  time.Now().UTC().Year(),
		ChainSlotsPerEpoch:    utils.Config.Chain.ClConfig.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.ClConfig.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		LatestFinalizedEpoch:  services.LatestFinalizedEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
		DepositContract:       utils.Config.Chain.ClConfig.DepositContractAddress,
		ChainConfig:           utils.Config.Chain.ClConfig,
		Debug:                 utils.Config.Frontend.Debug,
		GasNow:                services.LatestGasNowData(),
		ShowSyncingMessage:    services.IsSyncing(),
		GlobalNotification:    services.GlobalNotificationMessage(),
		MainMenuItems:         createMenuItems(active),
		TermsOfServiceUrl:     utils.Config.Frontend.Legal.TermsOfServiceUrl,
		PrivacyPolicyUrl:      utils.Config.Frontend.Legal.PrivacyPolicyUrl,
	}

	return data
}

func SetPageDataTitle(pageData *types.PageData, title string) {
	if title == "" {
		pageData.Meta.Title = fmt.Sprintf("%v - %v", utils.Config.Frontend.SiteName, time.Now().Year())
	} else {
		pageData.Meta.Title = fmt.Sprintf("%v - %v - %v", title, utils.Config.Frontend.SiteName, time.Now().Year())
	}
}

func createMenuItems(active string) []types.MainMenuItem {
	return []types.MainMenuItem{
		{
			Label:    "Blockchain",
			IsActive: active == "blockchain",
			Groups: []types.NavigationGroup{
				{
					Links: []types.NavigationLink{
						{
							Label: "Epochs",
							Path:  "/epochs",
							Icon:  "fa-history",
						},
						{
							Label: "Slots",
							Path:  "/slots",
							Icon:  "fa-cube",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Blocks",
							Path:  "/blocks",
							Icon:  "fa-cubes",
						},
						{
							Label: "Txs",
							Path:  "/transactions",
							Icon:  "fa-credit-card",
						},
						{
							Label: "Mempool",
							Path:  "/mempool",
							Icon:  "fa-upload",
						},
					},
				},
			},
		},
		{
			Label:    "Validators",
			IsActive: active == "validators",
			Groups: []types.NavigationGroup{
				{
					Links: []types.NavigationLink{
						{
							Label: "Overview",
							Path:  "/validators",
							Icon:  "fa-table",
						},
						{
							Label: "Slashings",
							Path:  "/validators/slashings",
							Icon:  "fa-user-slash",
						},
					},
				},
				{
					Links: []types.NavigationLink{

						{
							Label: "Deposits",
							Path:  "/validators/deposits",
							Icon:  "fa-file-signature",
						},
						{
							Label: "Withdrawals",
							Path:  "/validators/withdrawals",
							Icon:  "fa-money-bill",
						},
					},
				},
			},
		},
		{
			Label:    "Dashboard",
			IsActive: active == "dashboard",
			Path:     "/dashboard",
		},
		{
			Label:        "More",
			IsActive:     active == "more",
			HasBigGroups: true,
			Groups: []types.NavigationGroup{
				{
					Label: "Stats",
					Links: []types.NavigationLink{
						{
							Label: "Charts",
							Path:  "/charts",
							Icon:  "fa-chart-bar",
						},
						// TODO(now.youtrack.cloud/issue/TZB-13)
						/*
							{
								Label: "Income History",
								Path:  "/rewards",
								Icon:  "fa-money-bill-alt",
							},
						*/
						{
							Label: "Profit Calculator",
							Path:  "/calculator",
							Icon:  "fa-calculator",
						},
						{
							Label: "Block Viz",
							Path:  "/vis",
							Icon:  "fa-project-diagram",
						},
						{
							Label: "EIP-1559 Burn",
							Path:  "/burn",
							Icon:  "fa-burn",
						},
					},
				}, {
					Label: "Tools",
					Links: []types.NavigationLink{
						{
							Label: "Unit Converter",
							Path:  "/tools/unitConverter",
							Icon:  "fa-sync",
						},
						{
							Label: "GasNow",
							Path:  "/gasnow",
							Icon:  "fa-gas-pump",
						},
					},
				}, {
					Label: "Services",
					Links: []types.NavigationLink{
						{
							Label: "QRL Clients",
							Path:  "/qrlClients",
							Icon:  "fa-desktop",
						},
						{
							Label: "Slot Finder",
							Path:  "/slots/finder",
							Icon:  "fa-cube",
						},
					},
				},
			},
		},
	}
}
