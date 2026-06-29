package telegram

func playlistSize(seeds int) int {
	switch seeds {
	case 1:
		return 10
	case 2:
		return 15
	case 3:
		return 20
	default:
		return 30
	}
}
