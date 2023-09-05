"use strict";

class ColorDisplay {
  yourColor_;
  otherPlayerColor_;

  constructor(yourColor, otherPlayerColor) {
    this.yourColor_ = yourColor;
    this.otherPlayerColor_ = otherPlayerColor;
  }

  static clearColor(color) {
    while (color.lastChild) {
      color.removeChild(color.lastChild);
    }
  }
  clear() {
    ColorDisplay.clearColor(this.yourColor_);
    ColorDisplay.clearColor(this.otherPlayerColor_);
  }

  update(idToPlayer) {
    this.clear();
    for (const [id, player] of Object.entries(idToPlayer)) {
      if (hasCookieWithName(id)) {
        this.yourColor_.appendChild(document.createTextNode(player.color));
      } else {
        this.otherPlayerColor_.appendChild(
            document.createTextNode(player.color));
      }
    }
  }
}
