package meta

var localRegion string
var autoFailover bool
var leaderAddress string

func init() {
	localRegion = "bj"
	autoFailover = true
	leaderAddress = "http://127.0.0.1:6000"
}

func LocalRegion() string {
	return localRegion
}

func AutoFailover() bool {
	return autoFailover
}

func LeaderAddress() string {
	return leaderAddress
}
