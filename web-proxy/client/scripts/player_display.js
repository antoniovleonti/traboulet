"use strict";

class PlayerDisplay {
  clock_display_;
  color_;
  active_;
  score_;
  you_;

  constructor(params) {
    this.clock_display_ = new ClockDisplay(params.clock);
    this.color_ = params.color;
    this.active_ = params.active;
    this.score_ = params.score;
    this.you_ = params.you;
  }

  reset() {
    while (this.color_.lastChild) {
      this.color_.removeChild(this.color_.lastChild);
    }
    while (this.score_.lastChild) {
      this.score_.removeChild(this.score_.lastChild);
    }
  }

  update(info, isMe, isTheirTurn, timeControl, firstMoveDeadline) {
    console.log(info);
    this.reset();
    let isTheirFirstMove = (isTheirTurn && firstMoveDeadline != null);
    if (isTheirFirstMove) {
      info.deadline = firstMoveDeadline;
    }
    this.clock_display_.update(info.timeNs, timeControl, info.deadline);
    this.score_.appendChild(document.createTextNode(info.score));
    this.color_.appendChild(document.createTextNode(info.color));
    this.active_.hidden = !isTheirTurn;
    if (this.you_ != null) {
      this.you_.hidden = !isMe;
    }
  }
}
