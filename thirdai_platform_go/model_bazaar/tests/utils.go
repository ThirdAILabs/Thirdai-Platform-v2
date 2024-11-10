package tests

import (
	"slices"
	"thirdai_platform/model_bazaar/services"
)

func sortTeamList(users []services.TeamInfo) {
	slices.SortFunc(users, func(a, b services.TeamInfo) int {
		if a.Name == b.Name {
			return 0
		}
		if a.Name < b.Name {
			return -1
		}
		return 1
	})
}

func sortUserTeamList(users []services.UserTeamInfo) {
	slices.SortFunc(users, func(a, b services.UserTeamInfo) int {
		if a.TeamName == b.TeamName {
			return 0
		}
		if a.TeamName < b.TeamName {
			return -1
		}
		return 1
	})
}

func sortTeamUserList(users []services.TeamUserInfo) {
	slices.SortFunc(users, func(a, b services.TeamUserInfo) int {
		if a.Username == b.Username {
			return 0
		}
		if a.Username < b.Username {
			return -1
		}
		return 1
	})
}

func sortModelList(models []services.ModelInfo) {
	slices.SortFunc(models, func(a, b services.ModelInfo) int {
		if a.ModelName == b.ModelName {
			return 0
		}
		if a.ModelName < b.ModelName {
			return -1
		}
		return 1
	})
}

func sortUserList(users []services.UserInfo) {
	slices.SortFunc(users, func(a, b services.UserInfo) int {
		if a.Username == b.Username {
			return 0
		}
		if a.Username < b.Username {
			return -1
		}
		return 1
	})
}
