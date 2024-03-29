"use strict";

class HistoryManager {
  boardDisplay_;
  moveOl_;
  scrollWrapper_;

  currentSnapshotIdx_;
  history_;
  validMoves_; // (for current snapshot)
  myID_;
  isYourTurn_;

  constructor(boardDisplay, moveOl, scrollWrapper) {
    this.boardDisplay_ = boardDisplay;
    this.moveOl_ = moveOl;
    this.scrollWrapper_ = scrollWrapper;

    this.currentSnapshotIdx_ = 0;
    this.history_ = null;
    this.validMoves_ = null;

    this.myID_ = getMyID();
    this.isYourTurn_ = false;
  }

  clearMoveOl() {
    while (this.moveOl_.lastChild) {
      this.moveOl_.removeChild(this.moveOl_.lastChild);
    }
  }

  update(history, validMoves, idToPlayer) {
    this.history_ = history;
    this.validMoves_ = validMoves;

    console.log(this.myID_, idToPlayer[this.myID_]);
    this.isYourTurn_ = (
        (this.myID_ != null) &&
        (idToPlayer[this.myID_] != null) &&
        (idToPlayer[this.myID_].color ==
             this.history_[this.history_.length-1].whoseTurn));

    this.last();
  }

  updateMoveOl(history) {
    this.clearMoveOl();

    let selectedAnchor;
    let selectedLi;
    for (let i = 0; i < history.length; i++) {
      const move = history[i].lastMove;
      const li = document.createElement("li");
      const a = document.createElement('a');

      if (move == null) {
        a.appendChild(document.createTextNode('start'));
      } else {
        a.appendChild(document.createTextNode(
            "" + "ABCDEFG"[move.x] + (7 - move.y) + " " + move.d));
      }

      if (i == this.currentSnapshotIdx_) {
        selectedAnchor = a;
        selectedLi = li;
      } else {
        const this_ = this;
        a.addEventListener('click', (e) => {
          this_.at(i);
        });
      }

      li.appendChild(a);
      this.moveOl_.appendChild(li);
    }

    selectedAnchor.classList.add('bordered');
    selectedAnchor.classList.add('rounded');
    selectedAnchor.classList.add('padded-sm');
    selectedAnchor.classList.add('highlighted');

    // Scroll to selected element
    const offsetLeft = selectedLi.offsetLeft - this.moveOl_.offsetLeft;
    if (offsetLeft < this.scrollWrapper_.scrollLeft) {
      console.log("scrolling left...");
      this.scrollWrapper_.scrollTo({
        left: offsetLeft,
        behavior: "smooth",
      });
    }
    const offsetRight = offsetLeft + selectedLi.offsetWidth;
    if (offsetRight > this.scrollWrapper_.scrollLeft +
        this.scrollWrapper_.clientWidth) {
      console.log("scrolling right...");
      this.scrollWrapper_.scrollTo({
        left: offsetRight - this.scrollWrapper_.clientWidth,
        behavior: "smooth",
      });
    }
  }

  render() {
    const canMove = this.isYourTurn_ &&
                    (this.currentSnapshotIdx_ == this.history_.length-1);

    console.log(this.history_[this.currentSnapshotIdx_].board);
    this.boardDisplay_.update(this.history_[this.currentSnapshotIdx_].board,
                              this.validMoves_, canMove);

    this.updateMoveOl(this.history_, this.currentSnapshotIdx_);
  }

  next() {
    this.currentSnapshotIdx_ =
        Math.min(this.currentSnapshotIdx_ + 1, this.history_.length - 1)
    this.render();
  }

  prev() {
    this.currentSnapshotIdx_ = Math.max(this.currentSnapshotIdx_ - 1, 0)
    this.render();
  }

  first() {
    this.currentSnapshotIdx_ = 0;
    this.render();
  }

  last() {
    this.currentSnapshotIdx_ = this.history_.length - 1;
    this.render();
  }

  at(idx) {
    this.currentSnapshotIdx_ =
        Math.max(Math.min(idx, this.history_.length - 1), 0);
    this.render();
  }
}
