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

    let lastA = null;
    for (const snapshot of history) {
      if (snapshot.lastMove == null) {
        continue;
      }
      const move = snapshot.lastMove;
      const li = document.createElement("li");
      lastA = document.createElement('a');
      lastA.appendChild(document.createTextNode(
          "" + "ABCDEFG"[move.x] + (7 - move.y) + " " + move.d));
      li.appendChild(lastA);
      this.ol_.appendChild(li);
    }

    if (lastA != null) {
      lastA.classList.add('move-history--last-move');
    }

    if (this.ol_.lastChild != null) {
      this.ol_.lastChild.scrollIntoView(
          { inline: 'center', block: 'center' });
    } else {
      // append an empty li for proper height
      const li = document.createElement('li');
      li.appendChild(document.createTextNode('(no moves)'));
      this.ol_.appendChild(li);
    }
  }
}
