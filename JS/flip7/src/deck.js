function shuffleInPlace(array) {
  for (let i = array.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [array[i], array[j]] = [array[j], array[i]];
  }
}

function buildDeck() {
  // Distribution simple et intuitive:
  // 0 x1, 1 x2, 2 x3, ..., 12 x13
  const deck = [];
  for (let value = 0; value <= 12; value++) {
    const copies = value + 1;
    for (let k = 0; k < copies; k++) deck.push(value);
  }
  shuffleInPlace(deck);
  return deck;
}

class Deck {
  constructor() {
    this.cards = buildDeck();
    this.discard = [];
  }

  remaining() {
    return this.cards.length;
  }

  draw() {
    if (this.cards.length === 0) {
      // On remélange la défausse si besoin
      this.cards = this.discard;
      this.discard = [];
      shuffleInPlace(this.cards);
    }
    return this.cards.pop();
  }

  discardCard(card) {
    this.discard.push(card);
  }
}

module.exports = { Deck };
