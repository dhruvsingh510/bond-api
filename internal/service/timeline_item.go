package service

import (
	"context"
)

// TImeline Model
type TimelineItem struct {
	ID     int64 `json:"id,omitempty"`
	UserID int64 `json:"user_id"`
	PostID int64 `json:"post_id"`
	Post   Post  `json:"post,omitempty"`
}

type timelineItemClient struct {
	timeline chan TimelineItem
	userID int64
}

// SubscribeToTimeline to receive timeline items in realtime.
func (s * Service) SubscribeToTimeline(ctx context.Context) (chan TimelineItem, error) {
	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return nil, ErrUnauthenticated
	}

	tt := make(chan TimelineItem)
	c := &timelineItemClient{timeline: tt, userID: uid}
	s.timelineItemClients.Store(c, struct{}{})

	go func() {
		<-ctx.Done()
		s.timelineItemClients.Delete(c)
		close(tt)
	}()

	return tt, nil
}

func (s *Service) broadcastTimelineItem(ti TimelineItem) {
	s.timelineItemClients.Range(func(key, value interface{}) bool {
		c := key.(*timelineItemClient)
		if c.userID == ti.UserID {
			c.timeline <- ti
		}
		return true
	})
}