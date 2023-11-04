package evtpub

import (
  "net/url"
  "log"
)

type ChannelPublisher struct {
  channelURL *url.URL
  eventPub EventPublisher
}

func (scep *ChannelPublisher) init() error {
  log.Print("init channelURL: "+scep.channelURL.String())
  return scep.eventPub.initChannel(scep.channelURL)
}

func (scep *ChannelPublisher) Push(eventType, data string) error {
  log.Print("push channelURL: "+scep.channelURL.String())
  return scep.eventPub.push(scep.channelURL, eventType, data)
}

func (scep *ChannelPublisher) Delete() error {
  return scep.eventPub.deleteChannel(scep.channelURL)
}

