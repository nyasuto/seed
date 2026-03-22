// Package render provides ANSI-colored ASCII rendering for the chaosseed
// simulation. It extends core's plain-text rendering with:
//
//   - Element-based ANSI color coding (Fire=Red, Water=Blue, Wood=Green,
//     Metal=Yellow, Earth=Brown)
//   - Colored HP/progress bars
//   - Beast and invader detail listings
//   - Economy status display
//   - CoreHP bar visualization
//
// All output is designed to fit within 80 terminal columns of visible width.
package render
