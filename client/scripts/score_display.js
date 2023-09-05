"use strict";

class ScoreDisplay {
  yourScore_;
  otherPlayerScore_;

  constructor(yourScore, otherPlayerScore) {
    this.yourScore_ = yourScore;
    this.otherPlayerScore_ = otherPlayerScore;
  }

  static clearScore(score) {
    while (score.lastChild) {
      score.removeChild(score.lastChild);
    }
  }
  clear() {
    ScoreDisplay.clearScore(this.yourScore_);
    ScoreDisplay.clearScore(this.otherPlayerScore_);
  }

  update(idToPlayer) {
    this.clear();
    for (const [id, player] of Object.entries(idToPlayer)) {
      if (hasCookieWithName(id)) {
        this.yourScore_.appendChild(document.createTextNode(player.score));
      } else {
        this.otherPlayerScore_.appendChild(
            document.createTextNode(player.score));
      }
    }
  }
}
