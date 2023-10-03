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

  update(history, selectedIdx) {
    this.clear();

    let selectedAnchor;
    for (let i = 0; i < history.length; i++) {
      const move = history[i].lastMove;
      const li = document.createElement("li");
      const a = document.createElement('a');
      if (i == selectedIdx) {
        selectedAnchor = a;
      }
      if (move == null) {
        a.appendChild(document.createTextNode('start'));
      } else {
        a.appendChild(document.createTextNode(
            "" + "ABCDEFG"[move.x] + (7 - move.y) + " " + move.d));
      }
      li.appendChild(a);
      this.ol_.appendChild(li);
    }

    selectedAnchor.classList.add('move-history--last-move');
    selectedAnchor.scrollIntoView({ inline: 'center', block: 'center' });
  }
}
