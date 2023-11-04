"use strict";

class ClockDisplay {
  clock_;
  interval_;

  constructor(clock) {
    this.clock_ = clock;
  }

  reset() {
    clearInterval(this.interval_);
    ClockDisplay.clear(this.clock_);
    this.clock_.classList.remove("player-clock-active");
    this.clock_.classList.remove("player-clock");
    this.clock_.classList.add("player-clock");
  }

  static clear(clock) {
    while (clock.lastChild) {
      clock.removeChild(clock.lastChild);
    }
  }

  update(timeNs, timeControl, deadline) {
    this.reset()
    if (deadline != null) {
      this.interval_ = setInterval(
          ClockDisplay.tickClock, 10, this.clock_, Date.parse(deadline),
          timeControl);
    } else {
      ClockDisplay.writeClock(this.clock_, timeNs, timeControl);
    }
  }

	static writeClock(clock, durationNs, timeControl, ticking=false) {
    ClockDisplay.clear(clock);
    const fmt = ClockDisplay.formatNs(durationNs);
    const span = document.createElement('span');
    span.appendChild(document.createTextNode(fmt));
    clock.appendChild(span);
    if (ticking) {
      clock.classList.add("player-clock-active");
    }
  }

  static tickClock(clock, deadline, timeControl) {
    const durationMs = new Date(deadline - new Date());
    const durationNs = durationMs * 1e6;
    ClockDisplay.writeClock(
        clock, Math.max(durationNs, 0), timeControl, true);
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
