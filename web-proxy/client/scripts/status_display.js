"use strict";

class StatusDisplay {
  el_;

  constructor(el) {
    this.el_ = el;
  }

  clear() {
    while (this.el_.lastChild) {
      this.el_.removeChild(this.el_.lastChild);
    }
  }

  update(status) {
    this.clear();
    this.el_.appendChild(document.createTextNode(status));
  }
}
