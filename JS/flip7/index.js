const { startGame } = require("./src/game");

startGame().catch((err) => {
  console.error("Erreur fatale :", err);
  process.exit(1);
});
