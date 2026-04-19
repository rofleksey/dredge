package twitch

func (s *Usecase) StopMonitor() {
	s.live.StopMonitor()
}
