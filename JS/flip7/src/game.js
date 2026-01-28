// src/game.js
const path = require("path");
const { Deck } = require("./deck");
const { Logger } = require("./logger");
const { createPrompt, fileTimestamp, sum, uniqueCount } = require("./utils");

const TARGET_SCORE = 200;      // objectif
const FLIP7_BONUS = 15;        // bonus si 7 numéros différents dans la manche
const USE_FLIP7_BONUS = true;  // mets false si tu veux encore plus simple

function printRulesShort() {
  console.log("\n=== Flip7 (simplifié) - Mode texte ===");
  console.log("- Chaque joueur joue à tour de rôle (round-robin).");
  console.log("- À ton tour: tu choisis STOP (banquer) ou PIOCHER (1 carte).");
  console.log("- Si tu pioches un numéro déjà dans TA manche => BUST => 0 point pour la manche.");
  console.log("- Si tu STOP, tu ajoutes la somme de ta manche à ton score total et tu sors de la manche.");
  if (USE_FLIP7_BONUS) {
    console.log(
      `- Si tu as 7 numéros différents dans ta manche => FLIP 7 => +${FLIP7_BONUS} points et tu sors de la manche.`
    );
  }
  console.log(`- Fin de partie: dès qu'un joueur atteint ${TARGET_SCORE} points, la fin est DÉCLENCHÉE,`);
  console.log("  mais on termine la manche en cours. Le gagnant est celui qui a le meilleur score final.\n");
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

  // Fin "règle complète" : déclenchement puis fin de manche
  let endTriggered = false;
  let endTriggerInfo = null; // { roundIndex, player, score }

  while (true) {
    roundIndex++;

    // --- ÉTATS DE MANCHE (round-robin) ---
    // status: "ACTIVE" | "STOPPED" | "BUSTED" | "FLIP7"
    const roundState = new Map();
    for (const p of players) {
      roundState.set(p.name, { hand: [], status: "ACTIVE" });
    }

    const anyActive = () => {
      for (const p of players) {
        if (roundState.get(p.name).status === "ACTIVE") return true;
      }
      return false;
    };

    console.log(`\n========= Manche ${roundIndex} =========`);
    logger.log({ type: "ROUND_START", roundIndex });

    // Tant qu'il reste au moins un joueur actif, on fait tourner les tours
    while (anyActive()) {
      for (const player of players) {
        const st = roundState.get(player.name);
        if (st.status !== "ACTIVE") continue;

        console.log(`\n--- ${player.name} (score total: ${player.score}) ---`);
        console.log(`Main actuelle: ${formatHand(st.hand)}`);

        logger.log({
          type: "TURN_PROMPT",
          roundIndex,
          player: player.name,
          hand: st.hand.slice(),
          totalScoreBefore: player.score,
        });

        // Choix: Stop (banque maintenant) ou Piocher (1 carte)
        const pre = (await prompt.ask("Stop (s) ou Piocher (p) ? ")).trim().toLowerCase();

        if (pre.startsWith("s")) {
          const points = sum(st.hand);
          player.score += points;
          st.status = "STOPPED";

          logger.log({
            type: "BANK",
            roundIndex,
            player: player.name,
            hand: st.hand.slice(),
            pointsAdded: points,
            totalScoreAfter: player.score,
          });

          console.log(`STOP. Points ajoutés: ${points}. Nouveau score: ${player.score}`);

          // Déclenchement de fin (mais on ne stoppe pas tout de suite, on termine la manche)
          if (!endTriggered && player.score >= TARGET_SCORE) {
            endTriggered = true;
            endTriggerInfo = { roundIndex, player: player.name, score: player.score };

            logger.log({
              type: "END_TRIGGERED",
              roundIndex,
              player: player.name,
              totalScore: player.score,
              reason: "TARGET_REACHED_AFTER_BANK",
            });

            console.log(
              `\n⚑ Fin de partie déclenchée par ${player.name} (score: ${player.score}). On termine la manche ${roundIndex}...\n`
            );
          }

          continue;
        }

        // Sinon, il pioche UNE carte puis on passe au joueur suivant
        const card = deck.draw();

        logger.log({
          type: "DRAW",
          roundIndex,
          player: player.name,
          card,
          deckRemaining: deck.remaining(),
        });

        // BUST si doublon dans SA main
        if (st.hand.includes(card)) {
          deck.discardCard(card);
          st.status = "BUSTED";

          logger.log({
            type: "BUST",
            roundIndex,
            player: player.name,
            card,
            handBeforeBust: st.hand.slice(),
          });

          console.log(`Pioche: ${card} -> DOUBLON ! BUST. Manche = 0 point.`);
          continue;
        }

        st.hand.push(card);
        deck.discardCard(card);

        console.log(`Pioche: ${card} | Nouvelle main: ${formatHand(st.hand)}`);

        // FLIP 7 si 7 uniques
        if (USE_FLIP7_BONUS && uniqueCount(st.hand) >= 7) {
          const points = sum(st.hand) + FLIP7_BONUS;
          player.score += points;
          st.status = "FLIP7";

          logger.log({
            type: "FLIP7",
            roundIndex,
            player: player.name,
            hand: st.hand.slice(),
            basePoints: sum(st.hand),
            bonus: FLIP7_BONUS,
            pointsAdded: points,
            totalScoreAfter: player.score,
          });

          console.log(
            `FLIP 7 ! +${FLIP7_BONUS} bonus. Points ajoutés: ${points}. Nouveau score: ${player.score}`
          );

          // Déclenchement de fin (mais on ne stoppe pas tout de suite, on termine la manche)
          if (!endTriggered && player.score >= TARGET_SCORE) {
            endTriggered = true;
            endTriggerInfo = { roundIndex, player: player.name, score: player.score };

            logger.log({
              type: "END_TRIGGERED",
              roundIndex,
              player: player.name,
              totalScore: player.score,
              reason: "TARGET_REACHED_AFTER_FLIP7",
            });

            console.log(
              `\n⚑ Fin de partie déclenchée par ${player.name} (score: ${player.score}). On termine la manche ${roundIndex}...\n`
            );
          }
        }

        // Note: on ne demande pas "continuer/stop" après la pioche,
        // car le joueur décidera à son prochain passage.
      }
    }

    logger.log({
      type: "ROUND_END",
      roundIndex,
      scores: players.map((p) => ({ name: p.name, score: p.score })),
    });

    console.log("\nScores après la manche:");
    for (const p of players) console.log(`- ${p.name}: ${p.score}`);

    // Fin de partie APRES fin de manche, si déclenchée
    if (endTriggered) {
      const maxScore = Math.max(...players.map((p) => p.score));
      const winners = players.filter((p) => p.score === maxScore);

      logger.log({
        type: "GAME_END",
        roundIndex,
        endTriggeredBy: endTriggerInfo,
        finalScores: players.map((p) => ({ name: p.name, score: p.score })),
        winners: winners.map((w) => ({ name: w.name, score: w.score })),
        isTie: winners.length > 1,
      });

      console.log("\n==============================");
      if (winners.length === 1) {
        console.log(`🏁 FIN DE PARTIE ! Gagnant: ${winners[0].name} avec ${winners[0].score} points`);
      } else {
        console.log(
          `🏁 FIN DE PARTIE ! ÉGALITÉ à ${maxScore} points entre : ${winners.map((w) => w.name).join(", ")}`
        );
      }
      console.log("Scores finaux:");
      for (const p of players) console.log(`- ${p.name}: ${p.score}`);
      console.log("==============================\n");

      logger.close();
      prompt.close();
      return;
    }
  }
}

module.exports = { startGame };
