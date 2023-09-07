"use strict";

class PlayerDisplayManager {
  top_player_;
  bottom_player_;

  // top_player, bottom_player have fields:
  // { clock, color, active, score, you }
  constructor(top_player, bottom_player) {
    this.top_player_ =
        new PlayerDisplay(top_player.clock, top_player.color, top_player.active,
                          top_player.score, top_player.you);
    this.bottom_player_ =
        new PlayerDisplay(bottom_player.clock, bottom_player.color,
                          bottom_player.active, bottom_player.score,
                          bottom_player.you);
  }

  update(idToPlayer, colorToPlayer) {
    let myID = PlayerDisplayManager.getMyID(Object.keys(idToPlayer));
    if (myID == null) {
      this.top_player_.update(colorToPlayer["WHITE"], false);
      this.bottom_player_.update(colorToPlayer["BLACK"], false);
    } else {
      for (const [id, player] of Object.entries(idToPlayer)) {
        // Get other player.
        if (id == myID) {
          this.bottom_player_.update(player, true);
        } else {
          this.top_player_.update(player, false);
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
