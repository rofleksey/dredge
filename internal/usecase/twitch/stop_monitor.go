package twitch

func (s *Service) StopMonitor() {
	s.live.StopMonitor()
}
