"use strict";

class PlayerDisplay {
  clock_display_;
  color_;
  active_;
  score_;
  you_;

  constructor(clock, color, active, score, you) {
    this.clock_display_ = new ClockDisplay(clock);
    this.color_ = color;
    this.active_ = active;
    this.score_ = score;
    this.you_ = you;
  }

  reset() {
    while (this.color_.lastChild) {
      this.color_.removeChild(this.color_.lastChild);
    }
    while (this.score_.lastChild) {
      this.score_.removeChild(this.score_.lastChild);
    }
  }

  update(info, isMe) {
    this.reset();
    this.clock_display_.update(info.timeNs, info.deadline);
    this.score_.appendChild(document.createTextNode(info.score));
    this.color_.appendChild(document.createTextNode(info.color));
    this.active_.hidden = (info.deadline == null);
    if (this.you_ != null) {
      this.you_.hidden = !isMe;
    }
  }
}
