const path = require("path");
const { Deck } = require("./deck");
const { Logger } = require("./logger");
const { createPrompt, fileTimestamp, sum, uniqueCount } = require("./utils");

const TARGET_SCORE = 200;      // objectif
const FLIP7_BONUS = 15;        // bonus si 7 numéros différents dans la manche
const USE_FLIP7_BONUS = true;  // tu peux mettre false si tu veux encore plus simple

function printRulesShort() {
  console.log("\n=== Flip7 (simplifié) - Mode texte ===");
  console.log("- Chaque joueur joue à tour de rôle.");
  console.log("- À ton tour: tu pioches des cartes (numéros).");
  console.log("- Si tu repioches un numéro déjà dans TA manche => BUST => 0 point pour la manche.");
  console.log("- Sinon tu peux CONTINUER (c) ou STOP (s) pour 'bank' la somme de ta manche.");
  if (USE_FLIP7_BONUS) console.log(`- Si tu as 7 numéros différents dans ta manche => FLIP 7 => +${FLIP7_BONUS} points et manche terminée.`);
  console.log(`- Premier à atteindre ${TARGET_SCORE} points gagne.\n`);
}

function formatHand(hand) {
  return `[${hand.join(", ")}] (somme=${sum(hand)}, uniques=${uniqueCount(hand)})`;
}

async function startGame() {
  const prompt = createPrompt();
  printRulesShort();

  const nRaw = await prompt.ask("Nombre de joueurs (2-8) ? ");
  const n = Math.max(2, Math.min(8, parseInt(nRaw, 10) || 2));

  const players = [];
  for (let i = 0; i < n; i++) {
    const name = (await prompt.ask(`Nom du joueur ${i + 1} ? `)).trim() || `Joueur${i + 1}`;
    players.push({ name, score: 0 });
  }

  const logPath = path.join("logs", `game_${fileTimestamp()}.jsonl`);
  const logger = new Logger(logPath);

  logger.log({
    type: "GAME_START",
    players: players.map((p) => p.name),
    targetScore: TARGET_SCORE,
    flip7Bonus: USE_FLIP7_BONUS ? FLIP7_BONUS : 0,
  });

  console.log(`\nLog enregistré dans: ${logPath}\n`);

  const deck = new Deck();
  let roundIndex = 0;

  while (true) {
    roundIndex++;
    logger.log({ type: "ROUND_START", roundIndex });
    console.log(`\n========= Manche ${roundIndex} =========`);

    // Chaque joueur joue son tour de manche
    for (const player of players) {
      console.log(`\n--- Tour de ${player.name} (score total: ${player.score}) ---`);

      const hand = [];
      let busted = false;
      let finishedTurn = false;

      logger.log({
        type: "TURN_START",
        roundIndex,
        player: player.name,
        totalScoreBefore: player.score,
      });

      while (!finishedTurn) {
        const card = deck.draw();

        logger.log({
          type: "DRAW",
          roundIndex,
          player: player.name,
          card,
          deckRemaining: deck.remaining(),
        });

        // BUST si doublon dans la main du joueur
        if (hand.includes(card)) {
          busted = true;
          deck.discardCard(card);

          logger.log({
            type: "BUST",
            roundIndex,
            player: player.name,
            card,
            handBeforeBust: hand.slice(),
          });

          console.log(`Pioche: ${card} -> DOUBLON ! BUST. Manche = 0 point.`);
          finishedTurn = true;
          break;
        }

        hand.push(card);
        deck.discardCard(card);

        console.log(`Pioche: ${card} | Main: ${formatHand(hand)}`);

        // FLIP 7 si 7 uniques
        if (USE_FLIP7_BONUS && uniqueCount(hand) >= 7) {
          const points = sum(hand) + FLIP7_BONUS;

          player.score += points;

          logger.log({
            type: "FLIP7",
            roundIndex,
            player: player.name,
            hand: hand.slice(),
            basePoints: sum(hand),
            bonus: FLIP7_BONUS,
            pointsAdded: points,
            totalScoreAfter: player.score,
          });

          console.log(`FLIP 7 ! +${FLIP7_BONUS} bonus. Points ajoutés: ${points}. Nouveau score: ${player.score}`);
          finishedTurn = true;
          break;
        }

        // choix continue/stop
        const ans = (await prompt.ask("Continuer (c) ou Stop (s) ? ")).trim().toLowerCase();
        const choice = ans.startsWith("s") ? "STOP" : "CONTINUE";

        logger.log({
          type: "CHOICE",
          roundIndex,
          player: player.name,
          choice,
          hand: hand.slice(),
        });

        if (choice === "STOP") {
          const points = sum(hand);
          player.score += points;

          logger.log({
            type: "BANK",
            roundIndex,
            player: player.name,
            hand: hand.slice(),
            pointsAdded: points,
            totalScoreAfter: player.score,
          });

          console.log(`STOP. Points ajoutés: ${points}. Nouveau score: ${player.score}`);
          finishedTurn = true;
        } else {
          // continue, boucle
        }
      }

      logger.log({
        type: "TURN_END",
        roundIndex,
        player: player.name,
        busted,
        totalScoreAfter: player.score,
      });

      // condition de victoire immédiate
      if (player.score >= TARGET_SCORE) {
        logger.log({
          type: "GAME_END",
          winner: player.name,
          finalScores: players.map((p) => ({ name: p.name, score: p.score })),
          roundIndex,
        });

        console.log("\n==============================");
        console.log(`🏁 FIN DE PARTIE ! Gagnant: ${player.name} avec ${player.score} points`);
        console.log("Scores finaux:");
        for (const p of players) console.log(`- ${p.name}: ${p.score}`);
        console.log("==============================\n");

        logger.close();
        prompt.close();
        return;
      }
    }

    logger.log({
      type: "ROUND_END",
      roundIndex,
      scores: players.map((p) => ({ name: p.name, score: p.score })),
    });

    console.log("\nScores après la manche:");
    for (const p of players) console.log(`- ${p.name}: ${p.score}`);
  }
}

module.exports = { startGame };
