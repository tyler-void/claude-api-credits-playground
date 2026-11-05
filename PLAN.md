# Plan

Spend a monthly API credit on experiments that are fun to watch and produce real numbers: win rate, illegal-action rate, latency and cost per model.

## Ground rules (every experiment)

- **One key per workspace, with a spend limit** (see `docs/setup-keys.md`).
- **Every run has a `max_usd` cap**, and every game has a hard ceiling (3× its estimated cost). A run stops starting new games once spent plus in-flight estimates reach the cap.
- **Prices live in config.** Cost is estimated from the usage the API reports.
- **Stateless decisions:** one request per decision, with a cached prefix (tool + system) and one fresh user message carrying everything the player may know.
- **One tool, `act`,** with `action`, optional public `talk` (≤140 chars), and an optional private `memo` (≤280 chars).
- **Strikes:** an illegal or malformed action gets 2 retries, and the third failure forfeits. A timeout gets one retry, and a second timeout in a row forfeits. Provider errors are not strikes: back off up to 5 times, then abort with no result.
- **Fairness:** colours swap every game; games with luck use a duplicate format (each seed is played twice with seats swapped).
- **Logs:** every decision records its tokens (including cache reads), cost, latency and attempts. Every game must replay from its seed and action log.
- **Tests never call the real API.**

## Experiments

| # | Experiment | Status | Gate |
|---|---|---|---|
| 0 | Chess arena on an existing chess server | Built ([vibecheck-chess](https://github.com/tyler-void/vibecheck-chess) `cmd/arena`); gate partly run | 10 Haiku-vs-random games, then 4 Opus-vs-Haiku games with colours swapped, all finishing with no harness crash |
| 1 | Generic multi-seat game engine | Planned | Chess runs unchanged through a generic `/view` + `/act` API |
| 2 | Gomoku (15×15, five in a row, Vietnamese "both ends blocked" rule) | Planned | 20 mirrored games Haiku vs a heuristic bot |
| 3 | Shotgun duel (hidden shells, items, best of 3) | Planned | Information-boundary property test; 50 duplicate pairs Haiku vs an expected-value bot |
| 4 | Heads-up No-Limit Hold'em | Planned | Hand evaluator matches a reference on 10M hands; 500 duplicate hands Haiku vs a baseline bot |
| 5 | Leaderboard and replay viewer | Planned | Ratings per game with confidence intervals; ties shown when intervals overlap |

## Next steps

1. Finish the chess gate in one clean run (`arena.gate.yaml`, `max_usd: 15`). Report cost, latency and illegal-attempt rate per model.
2. Write up the Opus-vs-Haiku numbers in `docs/discoveries.md`.
3. Add a second key in its own workspace, and decide which experiment it funds.

## Budget rules of thumb

- Run bulk games with Haiku and Sonnet; save Opus and Fable for showcase matches.
- The monthly credit expires each billing cycle, so schedule big runs late in the cycle.
- Estimate first, then measure. Our early estimates were low on decisions per game.
