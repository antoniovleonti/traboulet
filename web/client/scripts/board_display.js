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
    const moves = BoardDisplay.createMoveMap(isYourTurn ? validMoves : []);
    this.renderBoardNoSelection(board, moves);
  }

  static createWhiteMarble() {
    const marble = document.createElement("div");
    marble.classList.add('marble-base');
    marble.classList.add('marble-white');
    return marble;
  }

  static createBlackMarble() {
    const marble = document.createElement("div");
    marble.classList.add('marble-base');
    marble.classList.add('marble-black');
    return marble;
  }

  static createRedMarble() {
    const marble = document.createElement("div");
    marble.classList.add('marble-base');
    marble.classList.add('marble-red');
    return marble;
  }

  static createNullMarble() {
    const marble = document.createElement("div");
    marble.classList.add('marble-base');
    marble.classList.add('marble-null');
    return marble;
  }

  static createMarbleFromString(s) {
    switch (s) {
      case " ":
        return BoardDisplay.createNullMarble();
      case "R":
        return BoardDisplay.createRedMarble();
      case "W":
        return BoardDisplay.createWhiteMarble();
      case "B":
        return BoardDisplay.createBlackMarble();
      default:
        return null;
    }
  }

  static createMoveMap(arr) {
    const movesMap = [];
    for (let y = 0; y < 7; y++) {
      movesMap.push([]);
      for (let x = 0; x < 7; x++) {
        movesMap[y].push([]);
      }
    }
    for (const move of arr) {
      movesMap[move.y][move.x].push(move.d);
    }
    return movesMap;
  }

  static postMove(move) {
    const urlParts = window.location.href.split("/");
    const gameID = urlParts[urlParts.length-1];

    console.log("sending move: ", move);
    fetch('/api/games/' + gameID + '/move',
          { method: 'POST', body: JSON.stringify(move) })
        .then(response => {
          if (!response.ok) {
            response.text().then(txt => {
              console.log(`${response.status} ${txt}`);
            });
          }
        });
  }

  static directionStrToDxDy(s) {
    switch (s) {
      case "UP":
        return { x: 0, y: -1 };
      case "DOWN":
        return { x: 0, y: 1 };
      case "LEFT":
        return { x: -1, y: 0 };
      case "RIGHT":
        return { x: 1, y: 0 };
      default:
        throw new Error("invalid diration");
    }
  }

  renderBoardNoSelection(board, validMoves) {
    this.clear();

    // header
    const marbles = [];
    for (let y = 0; y < board.length; y++) {
      marbles.push([]);
      for (let x = 0; x < board[y].length; x++) {
        const marble = BoardDisplay.createMarbleFromString(board[y][x]);

        if (validMoves[y][x].length > 0) {
          // Selection logic
          marble.classList.add('marble-selectable');
          marble.classList.add('marble-selectable');
          const this_ = this;
          marble.addEventListener('click', (e) => {
            const selection = { y: y, x: x };
            this_.renderBoardWithSelection(
                board, validMoves, selection);
          });
        }

        marbles[y].push(marble);
        this.el_.appendChild(marble);
      }
    }
  }

  renderBoardWithSelection(board, validMoves, selection) {
    this.clear();
    const this_ = this;

    // header
    const marbles = [];
    for (let y = 0; y < board.length; y++) {
      marbles.push([]);
      for (let x = 0; x < board[y].length; x++) {
        const marble = BoardDisplay.createMarbleFromString(board[y][x]);
        marbles[y].push(marble);
        this.el_.appendChild(marble);
      }
    }

    const selectedMarble = marbles[selection.y][selection.x];
    const validMovesFromSelection = validMoves[selection.y][selection.x];
    selectedMarble.classList.add('marble-selected');
    selectedMarble.classList.add('marble-selectable');
    selectedMarble.addEventListener('click', (e) => {
      this_.renderBoardNoSelection(board, validMoves);
    });

    for (const dirStr of validMovesFromSelection) {
      const d = BoardDisplay.directionStrToDxDy(dirStr);

      const moveDestination = marbles[selection.y + d.y][selection.x + d.x];
      moveDestination.classList.add('marble-selectable');
      // Add listener to make move
      moveDestination.addEventListener(
          'click', (e) => {
            BoardDisplay.postMove(
                { X: selection.x, Y: selection.y, D: dirStr });
          });
    }
  }
}
