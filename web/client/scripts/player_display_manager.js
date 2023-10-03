"use strict";

class PlayerDisplayManager {
  topPlayer_;
  bottomPlayer_;
  resignButton_;
  rematchButton_;
  rematchCounter_;

  // top_player, bottom_player have fields:
  // { clock, color, active, score, you }
  constructor(top_player, bottom_player, resign_button, rematch_button,
              rematch_counter) {
    this.topPlayer_ = new PlayerDisplay(top_player);
    this.bottomPlayer_ = new PlayerDisplay(bottom_player);
    this.resignButton_ = resign_button;
    this.rematchButton_ = rematch_button;
    this.rematchCounter_ = rematch_counter;
  }

  update(idToPlayer, colorToPlayer, whoseTurn, timeControl, firstMoveDeadline,
         status) {
    let myID = getMyID();

    if (myID == null) {
      this.resignButton_.hidden = true;
      this.rematchButton_.hidden = true;
      this.topPlayer_.update(colorToPlayer["WHITE"], false, timeControl);
      this.bottomPlayer_.update(colorToPlayer["BLACK"], false, timeControl);
    } else {
      let rematchCounterVal = 0;
      for (const [id, player] of Object.entries(idToPlayer)) {
        // Get other player.
        const isTheirTurn = player.color == whoseTurn;
        if (id == myID) {
          this.bottomPlayer_.update(
              player, true, isTheirTurn, timeControl, firstMoveDeadline);
        } else {
          this.topPlayer_.update(
              player, false, isTheirTurn, timeControl, firstMoveDeadline);
        }
        if (player.wantsRematch) {
          rematchCounterVal++;
        }
      }
      this.resignButton_.hidden = status != 'ONGOING';
      this.rematchButton_.hidden = status == 'ONGOING';
      while (this.rematchCounter_.lastChild) {
        this.rematchCounter_.removeChild(this.rematchCounter_.lastChild);
      }
      this.rematchCounter_.appendChild(
          document.createTextNode(""+rematchCounterVal));
    }
  }
}
