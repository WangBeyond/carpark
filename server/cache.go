package main

import "log"

type Cache struct {
	data []CarparkAvailability
	ok   bool
}

func (s *Server) updateCache(carparkList []CarparkAvailability) {
	log.Println("update cache")
	s.cache.Store(&Cache{data: carparkList, ok: true})
}

func (s *Server) getCache() ([]CarparkAvailability, bool) {
	data := s.cache.Load()
	if data == nil {
		return nil, false
	}

	log.Println("use cache")
	return data.(*Cache).data, data.(*Cache).ok
}

func (s *Server) invalidateCache() {
	log.Println("invalidate cache")
	s.cache.Store(&Cache{ok: false})
}
