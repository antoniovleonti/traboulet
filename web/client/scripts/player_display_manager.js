"use strict";

class PlayerDisplayManager {
  top_player_;
  bottom_player_;

  // top_player, bottom_player have fields:
  // { clock, color, active, score, you }
  constructor(top_player, bottom_player) {
    this.top_player_ = new PlayerDisplay(top_player);
    this.bottom_player_ = new PlayerDisplay(bottom_player);
  }

  update(idToPlayer, colorToPlayer, whoseTurn, timeControl, firstMoveDeadline) {
    let myID = PlayerDisplayManager.getMyID(Object.keys(idToPlayer));
    if (myID == null) {
      this.top_player_.update(colorToPlayer["WHITE"], false, timeControl);
      this.bottom_player_.update(colorToPlayer["BLACK"], false, timeControl);
    } else {
      for (const [id, player] of Object.entries(idToPlayer)) {
        // Get other player.
        const isTheirTurn = player.color == whoseTurn;
        if (id == myID) {
          this.bottom_player_.update(
              player, true, isTheirTurn, timeControl, firstMoveDeadline);
        } else {
          this.top_player_.update(
              player, false, isTheirTurn, timeControl, firstMoveDeadline);
        }
      }
    }
  }

  static getMyID(ids) {
    let cookies = document.cookie.split(";");
    for (const id of ids) {
      for (const cookie of cookies) {
        let nameval = cookie.split("=");
        if (nameval[0].trim() == id) {
          return id;
        }
      }
    }
    return null
  }
}
