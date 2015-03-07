package meta

var localRegion string
var autoFailover bool

func init() {
	localRegion = "bj"
	autoFailover = true
}

func LocalRegion() string {
	return localRegion
}

func AutoFailover() bool {
	return autoFailover
}
