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

  update(board, validMoves, isYourTurn) {
    this.renderBoardNoSelection(board, validMoves, isYourTurn);
  }

  static createMoveMap(arr) {
    let movesMap = {};
    for (const move of arr) {
      if (movesMap[move.y] == null) {
        movesMap[move.y] = {};
      }
      if (movesMap[move.y][move.x] == null) {
        movesMap[move.y][move.x] = []; // list of directions
      }
      movesMap[move.y][move.x].push(move.d);
    }
    return movesMap;
  }

  static createTextWithClass(text, classname) {
    const bg = document.createElement("span");
    bg.classList.add(classname);
    bg.appendChild(document.createTextNode(text));
    return bg
  }

  static createBoardBottomEdge() {
    const p = document.createElement("span");
    p.appendChild(document.createTextNode("   "));
    p.appendChild(
        BoardDisplay.createTextWithClass("+---------------+", "board-bg"));
    p.appendChild(document.createTextNode("\n   y"));
    return p
  }

  static createBoardLeftEdge(i) {
    const span = document.createElement("span");
    span.appendChild(document.createTextNode(" " + i + " "));
    span.appendChild(BoardDisplay.createTextWithClass("| ", "board-bg"));
    return span;
  }

  renderBoardNoSelection(board, validMoves, isYourTurn) {
    this.clear();
    const moves = BoardDisplay.createMoveMap(validMoves);

    // header
    let p = document.createElement("span");
    p.appendChild(BoardDisplay.createBoardHeader());

    for (let i = 0; i < board.length; i++) {
      p.appendChild(BoardDisplay.createBoardLeftEdge(i));
      for (let j = 0; j < board[i].length; j++) {
        const cellMoves = moves[i] == null ? null : moves[i][j];
        p.appendChild(this.createCellNoSelection(
            board[i][j], cellMoves, i, j, isYourTurn,
            {board: board, moves: validMoves}));
      }
      p.appendChild(BoardDisplay.createBoardRightEdge());
    }
    p.appendChild(BoardDisplay.createBoardBottomEdge());
    this.el_.appendChild(p);
  }

  static createMarble(marble, classOverride) {
    switch (marble) {
      case " ":
        return BoardDisplay.createTextWithClass(
            ".", classOverride != null ? classOverride : "marble-empty");
      case "R":
        return BoardDisplay.createTextWithClass(
            "@", classOverride != null ? classOverride : "marble-red");
      case "W":
        return BoardDisplay.createTextWithClass(
            "O", classOverride != null ? classOverride : "marble-o");
      case "B":
        return BoardDisplay.createTextWithClass(
            "X", classOverride != null ? classOverride : "marble-x");
      default:
        return null;
    }
  }

  static createBoardHeader() {
    const span = document.createElement("span");
    span.appendChild(document.createTextNode("     0 1 2 3 4 5 6\n   "));
    span.appendChild(
        BoardDisplay.createTextWithClass("+---------------+", "board-bg"));
    span.appendChild(document.createTextNode(" x\n"));
    return span;
  }

  createCellNoSelection(marble, moves, y, x, isYourTurn, cancelParams) {
    const txt = BoardDisplay.createMarble(marble);

    const cell = document.createElement("span");
    if (isYourTurn && moves != null) {
      // button
      const a = document.createElement("a");
      a.classList.add("select-marble");
      const this_ = this;
      a.addEventListener("click", e => {
        this_.renderBoardWithSelection(cancelParams.board, moves, y, x,
                                       cancelParams);
      }, false);

      a.appendChild(txt);
      cell.appendChild(a);
    } else {
      cell.appendChild(txt);
    }
    cell.appendChild(BoardDisplay.createTextWithClass(" ", "board-bg"));
    return cell
  }

  static createBoardRightEdge() {
    const span = document.createElement("span");
    span.appendChild(BoardDisplay.createTextWithClass("|", "board-bg"));
    span.appendChild(document.createTextNode("\n"));
    return span
  }

  // anchors at (x,y) to cancel and at moves
  renderBoardWithSelection(board, moves, y, x, cancelParams) {
    this.clear()
    const p = document.createElement("span");
    p.appendChild(BoardDisplay.createBoardHeader());

    for (let i = 0; i < board.length; i++) {
      p.appendChild(BoardDisplay.createBoardLeftEdge(i));
      for (let j = 0; j < board[i].length; j++) {
        p.appendChild(this.createCellWithSelection(
                          board[i][j], moves, i, j, y, x, cancelParams));
      }
      p.appendChild(BoardDisplay.createBoardRightEdge());
    }
    p.appendChild(BoardDisplay.createBoardBottomEdge());
    this.el_.appendChild(p);
  }

  createCellWithSelection(marble, moves, y, x, Y, X, cancelParams) {
    const cell = document.createElement("span");
    if (moves == null) {
      cell.appendChild(document.createTextNode(txt + " "));
      return cell
    }

    // Cell is button for up move
    if (moves.includes("UP") && x == X && y == Y-1) {
      const txt = BoardDisplay.createMarble(marble, "move-selection");
      cell.appendChild(this.createDestinationAnchor(txt, X, Y, "UP"));
      return cell;
    } else if (moves.includes("DOWN") && x == X && y == Y+1) {
      const txt = BoardDisplay.createMarble(marble, "move-selection");
      cell.appendChild(this.createDestinationAnchor(txt, X, Y, "DOWN"));
      return cell;
    } else if (moves.includes("RIGHT") && x == X+1 && y == Y) {
      const txt = BoardDisplay.createMarble(marble, "move-selection");
      cell.appendChild(this.createDestinationAnchor(txt, X, Y, "RIGHT"));
      return cell;
    } else if (moves.includes("LEFT") && x == X-1 && y == Y) {
      const txt = BoardDisplay.createMarble(marble, "move-selection");
      cell.appendChild(this.createDestinationAnchor(txt, X, Y, "LEFT"));
      return cell;
    } else if (x == X && y == Y) {
      const txt = BoardDisplay.createMarble(marble, "cancel-selection");
      cell.appendChild(this.createSourceCancelAnchor(txt, cancelParams));
      return cell
    } else {
      const txt = BoardDisplay.createMarble(marble);
      cell.appendChild(txt);
      cell.appendChild(BoardDisplay.createTextWithClass(" ", "board-bg"));
      return cell
    }
  }

  createSourceCancelAnchor(txt, cancelParams) {
    const span = document.createElement("span");
    const a = document.createElement("a");
    a.classList.add("cancel-selection");
    const this_ = this;
    a.addEventListener("click", e => {
      this_.renderBoardNoSelection(
          cancelParams.board, cancelParams.moves, true);
    }, false);

    a.appendChild(txt);
    span.appendChild(a);
    span.appendChild(BoardDisplay.createTextWithClass(" ", "board-bg"));
    return span;
  }

  createDestinationAnchor(txt, x, y, d) {
    const span = document.createElement("span");
    const a = document.createElement("a");
    a.classList.add("move-selection");
    a.addEventListener("click", e => {
      const m = JSON.stringify({
        x: x,
        y: y,
        d: d,
      });
      BoardDisplay.fetchPostMove(m);
    }, false);

    a.appendChild(txt);
    span.appendChild(a);
    span.appendChild(BoardDisplay.createTextWithClass(" ", "board-bg"));
    return span;
  }

  static fetchPostMove(move) {
    console.log("sending move: ", move);
    fetch(BoardDisplay.getURLBase() + '/move',
          { method: 'POST', body: move })
        .then(response => {
          if (!response.ok) {
            response.text().then(txt => {
              console.log(`${response.status} ${txt}`);
            });
          }
        });
  }

  static getURLBase() {
    let urlParts = window.location.href.split("/");
    let gameID = urlParts[urlParts.length-1];
    return '/api/games/' + gameID;
  }
}
