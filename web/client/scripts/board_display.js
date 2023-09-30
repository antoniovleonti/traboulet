"use strict";

class BoardDisplay {
  marbleLayer_;
  inputLayer_;

  constructor(marbleLayer, inputLayer) {
    this.marbleLayer_ = marbleLayer;
    this.inputLayer_ = inputLayer;
  }

  clear() {
    while (this.marbleLayer_.lastChild) {
      this.marbleLayer_.removeChild(this.marbleLayer_.lastChild);
    }
    while (this.inputLayer_.lastChild) {
      this.inputLayer_.removeChild(this.inputLayer_.lastChild);
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

  static createInputElement() {
    const el = document.createElement('div');
    el.classList.add('input-base');
    return el;
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
      movesMap[move.y][move.x].push({
        d: move.d,
        marblesMoved: move.marblesMoved,
      });
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
    for (let y = 0; y < board.length; y++) {
      for (let x = 0; x < board[y].length; x++) {
        const marble = BoardDisplay.createMarbleFromString(board[y][x]);
        const input = BoardDisplay.createInputElement();

        if (validMoves[y][x].length > 0) {
          // Selection logic
          input.classList.add('marble-selectable');
          const this_ = this;
          input.addEventListener('click', (e) => {
            const selection = { y: y, x: x };
            this_.renderBoardWithSelection(
                board, validMoves, selection);
          });
        }

        this.marbleLayer_.appendChild(marble);
        this.inputLayer_.appendChild(input);
      }
    }
  }

  renderBoardWithSelection(board, validMoves, selection) {
    this.clear();

    // header
    const marbles = [];
    const inputs = [];
    for (let y = 0; y < board.length; y++) {
      marbles.push([]);
      inputs.push([]);
      for (let x = 0; x < board[y].length; x++) {
        const marble = BoardDisplay.createMarbleFromString(board[y][x]);
        marbles[y].push(marble);
        this.marbleLayer_.appendChild(marble);

        const input = BoardDisplay.createInputElement();
        inputs[y].push(input);
        this.inputLayer_.appendChild(input);
      }
    }

    const selectedMarble = marbles[selection.y][selection.x];
    const selectedInput = inputs[selection.y][selection.x];
    const validMovesFromSelection = validMoves[selection.y][selection.x];
    selectedMarble.classList.add('marble-selected');
    selectedInput.classList.add('marble-selectable');

    const this_ = this;
    selectedInput.addEventListener('click', (e) => {
      this_.renderBoardNoSelection(board, validMoves);
    });

    for (const move of validMovesFromSelection) {
      const d = BoardDisplay.directionStrToDxDy(move.d);

      let x = selection.x;
      let y = selection.y;
      for (let diff = 0; diff < Math.max(1, move.marblesMoved - 1); diff++) {
        x += d.x;
        y += d.y;
        if (x < 0 || x >= 7 || y < 0 || y >= 7) {
          break;
        }
        inputs[y][x].classList.add('marble-selectable');
        inputs[y][x].addEventListener('click', (e) => {
          BoardDisplay.postMove({ X: selection.x, Y: selection.y, D: move.d });
        });
      }
    }
  }
}
