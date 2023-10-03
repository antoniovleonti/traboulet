"use strict";

class HistoryManager {
  boardDisplay_;
  moveListDisplay_;

  currentSnapshotIdx_;
  history_;
  validMoves_; // (for current snapshot)
  myID_;
  isYourTurn_;

  constructor(boardDisplay, moveListDisplay) {
    this.boardDisplay_ = boardDisplay;
    this.moveListDisplay_ = moveListDisplay;

    this.currentSnapshotIdx_ = 0;
    this.history_ = null;
    this.validMoves_ = null;

    this.myID_ = getMyID();
    this.isYourTurn_ = false;
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
    this.render();
  }

  render() {
    console.log(this.history_.length - 1, this.currentSnapshotIdx_);
    const canMove =
        this.isYourTurn_ &&
        (this.currentSnapshotIdx_ == this.history_.length-1);
    console.log(canMove);
    this.boardDisplay_.update(this.history_[this.currentSnapshotIdx_].board,
                              this.validMoves_, canMove);

    this.moveListDisplay_.update(this.history_);
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
}
