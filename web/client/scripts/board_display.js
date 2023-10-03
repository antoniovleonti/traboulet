"use strict";

class BoardDisplay {
  marbleLayer_;
  inputLayer_;

  constructor(marbleLayer, inputLayer) {
    this.marbleLayer_ = marbleLayer;
    this.inputLayer_ = inputLayer;
  }

  clearMarbles() {
    while (this.marbleLayer_.lastChild) {
      this.marbleLayer_.removeChild(this.marbleLayer_.lastChild);
    }
  }

  clear() {
    this.clearMarbles();
    while (this.inputLayer_.lastChild) {
      this.inputLayer_.removeChild(this.inputLayer_.lastChild);
    }
  }

  update(board, validMoves, isYourTurn) {
    const moves = BoardDisplay.createMoveMap(isYourTurn ? validMoves : []);
    this.renderNoSelection(board, moves);
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
        throw new Error("Invalid direction " + s + "!");
    }
  }

  renderMarblesNoPreview(board) {
    this.clearMarbles();

    for (let y = 0; y < board.length; y++) {
      for (let x = 0; x < board[y].length; x++) {
        const marble = BoardDisplay.createMarbleFromString(board[y][x]);
        this.marbleLayer_.appendChild(marble);
      }
    }
  }

  renderMarblesWithPreview(board, selection, move) {
    this.clearMarbles();

    const moved = [];
    for (let y = 0; y < board.length; y++) {
      moved.push([]);
      for (let x = 0; x < board[y].length; x++) {
        moved[y].push(false);
      }
    }
    // don't edit passed array
    const d = BoardDisplay.directionStrToDxDy(move.d);
    // edit board to include move
    board = structuredClone(board);
    let marblesMoved = 0;
    let tmp = ' ';
    let x = selection.x;
    let y = selection.y;
    for (; y < 7 && y >= 0 && x < 7 && y >= 0;) {
      [ board[y][x], tmp ] = [ tmp, board[y][x] ];
      moved[y][x] = true;
      if (marblesMoved == move.marblesMoved) {
        break;
      }
      marblesMoved++;
      x += d.x;
      y += d.y;
    }

    const marbles = []
    for (let y = 0; y < board.length; y++) {
      marbles.push([]);
      for (let x = 0; x < board[y].length; x++) {
        const marble = BoardDisplay.createMarbleFromString(board[y][x]);
        if (moved[y][x]) {
          marble.classList.add('marble-ghost');
        }
        marbles[y].push(marble);
        this.marbleLayer_.appendChild(marble);
      }
    }
  }

  renderNoSelection(board, validMoves) {
    this.clear();

    this.renderMarblesNoPreview(board);
    // header
    for (let y = 0; y < board.length; y++) {
      for (let x = 0; x < board[y].length; x++) {
        const input = BoardDisplay.createInputElement();

        if (validMoves[y][x].length > 0) {
          // Selection logic
          input.classList.add('input-selectable');
          const this_ = this;
          input.addEventListener('click', (e) => {
            const selection = { y: y, x: x };
            this_.renderSelection(
                board, validMoves, selection);
          });
        }

        this.inputLayer_.appendChild(input);
      }
    }
  }

  renderSelection(board, validMoves, selection) {
    this.clear();

    const this_ = this;

    this.renderMarblesNoPreview(board);
    // header
    const inputs = [];
    const isPreviewListener = [];
    for (let y = 0; y < board.length; y++) {
      inputs.push([]);
      isPreviewListener.push([]);
      for (let x = 0; x < board[y].length; x++) {
        const input = BoardDisplay.createInputElement();
        inputs[y].push(input);
        this.inputLayer_.appendChild(input);

        isPreviewListener[y].push(false);
      }
    }

    const selectedInput = inputs[selection.y][selection.x];
    const validMovesFromSelection = validMoves[selection.y][selection.x];
    selectedInput.classList.add('input-selected');

    selectedInput.addEventListener('click', (e) => {
      this_.renderNoSelection(board, validMoves);
    });

    for (const move of validMovesFromSelection) {
      const d = BoardDisplay.directionStrToDxDy(move.d);

      let x = selection.x;
      let y = selection.y;
      // For each valid move from selection add click & hover listeners.
      for (let diff = 0; diff < move.marblesMoved; diff++) {
        x += d.x;
        y += d.y;
        if (x < 0 || x >= 7 || y < 0 || y >= 7) {
          break;
        }
        isPreviewListener[y][x] = true;
        inputs[y][x].classList.add('input-selectable');
        inputs[y][x].addEventListener('click', (e) => {
          BoardDisplay.postMove({ X: selection.x, Y: selection.y, D: move.d });
        });
        inputs[y][x].addEventListener('mouseover', (e) => {
          this_.renderMarblesWithPreview(board, selection, move);
        });
      }
    }

    for (let y = 0; y < inputs.length; y++) {
      for (let x = 0; x < inputs[y].length; x++) {
        if (!isPreviewListener[y][x]) {
          inputs[y][x].addEventListener('mouseover', (e) => {
            this_.renderMarblesNoPreview(board);
          });
          inputs[y][x].addEventListener('click', (e) => {
            this_.renderNoSelection(board, validMoves);
          });
        }
      }
    }
    inputs[selection.y][selection.x].addEventListener('mouseover', (e) => {
      this_.renderMarblesNoPreview(board);
    });
  }
}
