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

  update(board, validMoves) {
    this.clear();
    this.el_.appendChild(this.getTextNodeFromBoard(board, validMoves));
  }

  getTextNodeFromBoard(board, validMoves) {
    let p = document.createElement("span");
    p.appendChild(document.createTextNode(
        "     0 1 2 3 4 5 6\n   +---------------+ x\n"));
    for (let i = 0; i < board.length; i++) {
      p.appendChild(document.createTextNode(" " + i + " | "));
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

        let hasValidMove = false;
        for (const move of validMoves) {
          if (j == move.x && i == move.y) {
            hasValidMove = true;
            break;
          }
        }
        let el;
        if (hasValidMove) {
          el = document.createElement("a");
          el.addEventListener("click", el => {
            console.log("click ", j, ", ", i);
          }, false);
          el.appendChild(document.createTextNode(c + " "));
          p.appendChild(el);
        } else {
          p.appendChild(document.createTextNode(c + " "));
        }
      }
      p.appendChild(document.createTextNode("|\n"));
    }
    p.appendChild(document.createTextNode("   +---------------+\n   y"));
    return p;
  }
}
