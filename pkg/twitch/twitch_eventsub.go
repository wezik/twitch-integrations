package twitch

import (
	"log"

	"github.com/nicklaw5/helix/v2"
)

func (c *TwitchConn) SubscribeToCustomRewards() {
	log.Println("Subscribing to custom rewards")
	res, err := c.AppClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: c.Broadcaster.ID,
			UserID:            c.Broadcaster.ID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://localhost:443/eventsub",
			Secret:   "1234567890",
		},
	})
	if err != nil {
		log.Println(err)
	}
	log.Printf("Subscription status: %d\n", res.StatusCode)
}
