"use strict";

class MoveHistoryDisplay {
  ol_;

  constructor(ol) {
    this.ol_ = ol;
  }

  clear() {
    while (this.ol_.lastChild) {
      this.ol_.removeChild(this.ol_.lastChild);
    }
  }

  update(history) {
    this.clear();

    for (const snapshot of history) {
      if (snapshot.lastMove == null) {
        continue;
      }
      const move = snapshot.lastMove;
      const li = document.createElement("li");
      li.appendChild(document.createTextNode(
          "" + "ABCDEFG"[move.x] + (7 - move.y) + " " + move.d));
      this.ol_.appendChild(li);
    }
  }
}
