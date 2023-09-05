"use strict";

class BoardDisplay {
  el_;

  constructor(el) {
    this.el_ = el;
  }

  clear() {
    while (this.el_.lastChild) {
      this.el_.removeChild(this.el_.lastChild);
    }
  }

  update(board) {
    this.clear();
    this.el_.appendChild(this.getTextNodeFromBoard(board));
  }

  getTextNodeFromBoard(board) {
    let t = "     0 1 2 3 4 5 6\n   +---------------+ x\n";
    for (let i = 0; i < board.length; i++) {
      t += " " + i + " | ";
      for (let j = 0; j < board[i].length; j++) {
        let c;
        switch (board[i][j]) {
          case " ":
            c = ".";
            break;
          case "R":
            c = "@";
            break;
          default:
            c = board[i][j];
            break;
        }
        t += c + " ";
      }
      t += "|\n";
    }
    t += "   +---------------+\n   y";
    return document.createTextNode(t);
  }
}
