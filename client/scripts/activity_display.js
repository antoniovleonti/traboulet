"use strict";

class ActivityDisplay {
  yourActivity_;
  otherPlayerActivity_;

  constructor(yourActivity, otherPlayerActivity) {
    this.yourActivity_ = yourActivity;
    this.otherPlayerActivity_ = otherPlayerActivity;
  }

  update(whoseTurn, colorToPlayer) {
    let youAreActive = hasCookieWithName(colorToPlayer[whoseTurn].id);
    this.yourActivity_.hidden = !youAreActive;
    this.otherPlayerActivity_.hidden = youAreActive;
  }
}
