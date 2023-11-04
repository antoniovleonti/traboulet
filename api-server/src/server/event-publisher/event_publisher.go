package evtpub

import (
  "io"
  "strings"
  "errors"
  "net/http"
  "net/url"
  "sync"
)

type EventPublisher interface {
  initChannel(path *url.URL) error
  push(path *url.URL, eventType, data string) error
  deleteChannel(path *url.URL) error
  NewChannelPublisher(channelPath string) (*ChannelPublisher, error)
}

type ExtEventPublisher struct {
  hostname string
  client http.Client
  channels map[string]struct{} // Active (publishable) channels
  mutex sync.RWMutex
}

func NewExtEventPublisher(hostname string) *ExtEventPublisher {
  return &ExtEventPublisher{
    hostname: hostname,
    channels: make(map[string]struct{}),
  }
}

func (pp ExtEventPublisher) checkChannelAndSendRequest(
  method string, path *url.URL, eventType string, data string,
  channelShouldExist bool) error {

  _, ok := pp.channels[path.String()]
  if channelShouldExist && !ok {
    return errors.New("No stream exists at '"+path.String()+"'.")
  } else if !channelShouldExist && ok {
    return errors.New("Stream already exists at '"+path.String()+"'.")
  }

  endpoint, err :=
    url.JoinPath(pp.hostname, path.String(), "event-publisher")

  if err != nil {
    return err
  }
  req, err := http.NewRequest(
    method, endpoint, strings.NewReader(data))
  if err != nil {
    return err
  }


  if eventType != "" {
    req.Header.Set("event-type", eventType)
  }

  resp, err := pp.client.Do(req)
  if err != nil {
    return err
  }

  if resp.StatusCode != http.StatusOK {
    b, _ := io.ReadAll(resp.Body)
    return errors.New("Error from proxy: '"+string(b)+"'.")
  }
  return nil
}

func (pp *ExtEventPublisher) initChannel(path *url.URL) error {
  pp.mutex.Lock()
  defer pp.mutex.Unlock()

  err := pp.checkChannelAndSendRequest(
    "POST", path, "init", "Initializing channel", false)
  if err != nil {
    return err
  }

  pp.channels[path.String()] = struct{}{}

  return nil
}

func (pp *ExtEventPublisher) push(
  path *url.URL, eventType, data string) error {

  pp.mutex.RLock()
  defer pp.mutex.RUnlock()

  err := pp.checkChannelAndSendRequest("POST", path, eventType, data, true)
  if err != nil {
    return err
  }

  return nil
}

func (pp *ExtEventPublisher) deleteChannel(path *url.URL) error {
  pp.mutex.Lock()
  defer pp.mutex.Unlock()

  err := pp.checkChannelAndSendRequest("DELETE", path, "", "", true)
  if err != nil {
    return err
  }

  delete(pp.channels, path.String())

  return nil
}

func (pp *ExtEventPublisher) NewChannelPublisher(
  channelPath string) (*ChannelPublisher, error) {

  parsedPath, err := url.Parse(channelPath)
  if err != nil {
    return nil, err
  }

  cp := ChannelPublisher{ parsedPath, pp }
  if err = cp.init(); err != nil {
    return nil, err
  }

  return &cp, nil
}

type MockEvent struct {
  Event string
  Data string
}

type MockStream struct {
  Pushes []MockEvent
  Deleted bool
}

type MockEventPublisher struct {
  // MockEventPublisher simply saves readers for tests to consume later.
  Channels map[string]*MockStream
}

func NewMockEventPublisher() *MockEventPublisher {
  return &MockEventPublisher{
    Channels: make(map[string]*MockStream),
  }
}

func (mp *MockEventPublisher) initChannel(path *url.URL) error {
  if _, ok := mp.Channels[path.String()]; ok {
    return errors.New("Stream already exists at '"+path.String()+"'.")
  }
  mp.Channels[path.String()] = &MockStream{}
  return nil
}

func (mp *MockEventPublisher) push(
  path *url.URL, eventType, data string) error {

  if _, ok := mp.Channels[path.String()]; !ok {
    return errors.New("No stream exists at '"+path.String()+"'.")
  }
  mp.Channels[path.String()].Pushes =
    append(mp.Channels[path.String()].Pushes, MockEvent{
      Event: eventType,
      Data: data,
    })
  return nil
}

func (mp *MockEventPublisher) deleteChannel(path *url.URL) error {
  stream, ok := mp.Channels[path.String()]
  if !ok {
    return errors.New("No stream exists at '"+path.String()+"'.")
  }
  if stream.Deleted {
    return errors.New("Stream at '"+path.String()+"' was already deleted.")
  }
  // Don't delete just in case a test wants to check what was pushed.
  stream.Deleted = true
  return nil
}

func (mp *MockEventPublisher) NewChannelPublisher(
  channelPath string) (*ChannelPublisher, error) {

  parsedPath, err := url.Parse(channelPath)
  if err != nil {
    return nil, err
  }

  cp := ChannelPublisher{ parsedPath, mp }
  if err = cp.init(); err != nil {
    return nil, err
  }

  return &cp, nil
}
