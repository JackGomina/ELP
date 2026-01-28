const fs = require("fs");
const path = require("path");
const { nowIso } = require("./utils");

class Logger {
  constructor(logFilePath) {
    this.logFilePath = logFilePath;
    const dir = path.dirname(logFilePath);
    fs.mkdirSync(dir, { recursive: true });
    this.stream = fs.createWriteStream(logFilePath, { flags: "a", encoding: "utf-8" });
  }

  log(event) {
    const payload = {
      ts: nowIso(),
      ...event,
    };
    this.stream.write(JSON.stringify(payload) + "\n");
  }

  close() {
    this.stream.end();
  }
}

module.exports = { Logger };
