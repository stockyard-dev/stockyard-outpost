package server

type Tier string

const (
	TierFree Tier = "free"
	TierPro  Tier = "pro"
)

type Limits struct {
	Tier        Tier
	Description string
}

func LimitsFor(tier string) Limits {
	if tier == "pro" {
		return Limits{Tier: TierPro, Description: "Unlimited widgets and habits"}
	}
	return Limits{Tier: TierFree, Description: "5 widgets, 3 habits"}
}

func (l Limits) IsPro() bool {
	return l.Tier == TierPro
}
