package consts

import "lakelens/internal/dto"

var SearchChoices = []dto.SearchChoice{
	{
		Label: "New Lake",
		Link:  "/newlake",
	},
}

var LatestFeaturesTour int32 = 1 // should we greater than 0, always.
