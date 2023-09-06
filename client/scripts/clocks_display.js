"use strict";

class ClocksDisplay {
  yourClock_;
  otherPlayerClock_;
  interval_;

  constructor(yourClock, otherPlayerClock) {
    this.yourClock_ = yourClock;
    this.otherPlayerClock_ = otherPlayerClock;
  }

  static clear(clock) {
    while (clock.lastChild) {
      clock.removeChild(clock.lastChild);
    }
  }

  update(idToPlayer) {
    clearInterval(this.interval_);
    for (const [id, player] of Object.entries(idToPlayer)) {
      if (hasCookieWithName(id)) {
        this.updateYourClock(player);
      } else {
        this.updateOtherPlayerClock(player);
      }
    }
  }

  updateYourClock(player) {
    this.updateClock(this.yourClock_, player);
  }

  updateOtherPlayerClock(player) {
    this.updateClock(this.otherPlayerClock_, player);
  }

  updateClock(clock, player) {
    if (player.deadline != null) {
      const deadline = Date.parse(player.deadline);
      this.interval_ =
          setInterval(ClocksDisplay.tickClock, 10, clock, deadline);
    } else {
      ClocksDisplay.writeClock(clock, player.timeNs);
    }
  }

	static writeClock(clock, durationNs) {
    ClocksDisplay.clear(clock);
    const fmt = ClocksDisplay.formatNs(durationNs);
    clock.appendChild(document.createTextNode(fmt));
  }

  static tickClock(clock, deadline) {
    const durationMs = new Date(deadline - new Date());
    const durationNs = durationMs * 1e6;
    ClocksDisplay.writeClock(clock, durationNs);
  }

  static formatNs(ns) {
    const nsPerSec = 1e9;
    const nsPerMin = 6e10;
    const nsPerHour = 36e11;

    const hours = Math.floor(ns / nsPerHour);
    const minutes = Math.floor((ns - hours * nsPerHour) / nsPerMin);
    const seconds =
        Math.floor((ns - hours * nsPerHour - minutes * nsPerMin) / nsPerSec);

    const hoursPadded = hours.toString().padStart(2, '0');
    const minutesPadded = minutes.toString().padStart(2, '0');
    const secondsPadded = seconds.toString().padStart(2, '0');

    return (hours > 0 ? hoursPadded + ":" : "") +
           minutesPadded + ":" + secondsPadded;
  }
}

function hasCookieWithName(name) {
  let cookies = document.cookie.split(";");
  for (let i = 0; i < cookies.length; i++) {
    let nameval = cookies[i].split("=");
    if (nameval[0].trim() == name) {
      return true
    }
  }
  return false
}
