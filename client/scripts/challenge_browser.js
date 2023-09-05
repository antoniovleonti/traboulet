"use strict";

class ChallengeBrowser {
  table_;

  constructor(table) {
    this.table_ = table;
  }

  clear() {
    while (this.table_.lastChild) {
      this.table_.removeChild(this.table_.lastChild);
    }
  }

  update(challenges) {
    this.clear();
    // `challenges` is a map of [id]->{challenge}
    // a challenge has 1 field:
    // config
    // which itself has 1 field:
    // initialTimeNs
    if (JSON.stringify(challenges) == '{}') {
      this.table_.appendChild(this.createEmpty());
    }
    for (const [id, challenge] of Object.entries(challenges)) {
      this.table_.appendChild(this.createRow(id, challenge));
		}
  }

  createEmpty() {
    let tr = document.createElement("tr");
    let td = document.createElement("td");
    td.colSpan = "2";
    let txt = document.createTextNode("no games found.");
    tr.appendChild(td);
    td.appendChild(txt);
    return tr;
  }

	createRow(id, challenge) {
    let tr = document.createElement("tr");
    let td = document.createElement("td");
    let ns = challenge.config.timeControlNs;
    let timeTxt = document.createTextNode(ChallengeBrowser.formatNs(ns));
    tr.appendChild(td);
    td.appendChild(timeTxt);

    let a = document.createElement("a");
    a.title = "You will be brought to another page to confirm.";
    a.href = "/challenges/" + id;
    let atxt = document.createTextNode("View challenge");
    tr.appendChild(a);
    a.appendChild(atxt);

    return tr;
  }

  static formatNs(ns) {
    const nsPerSec = 1e9;
    const nsPerMin = 6e10;
    const nsPerHour = 36e11;

    const hours = Math.floor(ns / nsPerHour);
    const minutes = Math.floor((ns - hours * nsPerHour) / nsPerMin);
    const seconds =
        Math.floor((ns - hours * nsPerHour - minutes * nsPerMin) / nsPerSec);

    return (hours > 0 ? hours.toString() + "h " : "") +
           (minutes > 0 ? minutes.toString() + "m " : "") +
           (seconds > 0 ? seconds.toString() + "s " : "");
  }
}
