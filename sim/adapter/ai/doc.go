// Package ai implements the AI Mode adapter for chaosseed-sim.
// It provides a JSON Lines protocol over stdin/stdout for external
// AI agents to interact with the game engine programmatically.
//
// Protocol overview:
//   - Server sends one JSON object per line to stdout
//   - Client sends one JSON object per line via stdin
//   - Message types: "state", "game_end", "error" (server → client),
//     "action" (client → server)
package ai
