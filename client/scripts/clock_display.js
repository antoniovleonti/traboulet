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

	static writeClock(clock, durationNs, timeControl) {
    ClockDisplay.clear(clock);
    const fmt = ClockDisplay.formatNs(durationNs);
    clock.appendChild(document.createTextNode(fmt));
  }

  static tickClock(clock, deadline, timeControl) {
    const durationMs = new Date(deadline - new Date());
    const durationNs = durationMs * 1e6;
    ClockDisplay.writeClock(
        clock, Math.max(durationNs, 0), timeControl);
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
