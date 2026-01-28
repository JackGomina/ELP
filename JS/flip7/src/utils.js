const readline = require("readline");

function createPrompt() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  function ask(question) {
    return new Promise((resolve) => rl.question(question, (ans) => resolve(ans)));
  }

  function close() {
    rl.close();
  }

  return { ask, close };
}

function nowIso() {
  return new Date().toISOString();
}

function pad2(n) {
  return String(n).padStart(2, "0");
}

function fileTimestamp() {
  const d = new Date();
  return (
    d.getFullYear() +
    pad2(d.getMonth() + 1) +
    pad2(d.getDate()) +
    "_" +
    pad2(d.getHours()) +
    pad2(d.getMinutes()) +
    pad2(d.getSeconds())
  );
}

function sum(arr) {
  return arr.reduce((a, b) => a + b, 0);
}

function uniqueCount(arr) {
  return new Set(arr).size;
}

module.exports = {
  createPrompt,
  nowIso,
  fileTimestamp,
  sum,
  uniqueCount,
};
