import React, { Component } from 'react';

export default class Grid extends Component {
  render() {
    const { size, robots } = this.props;

    let visionCells = {};

    const arrows = [[0, -1], [1, 0], [0, 1], [-1, 0]];

    robots.forEach(r => {
      switch (r.vision) {
        case 8:
          visionCells[`${r.x - 1}-${r.y - 1}`] = true;
          visionCells[`${r.x - 1}-${r.y}`] = true;
          visionCells[`${r.x - 1}-${r.y + 1}`] = true;
          visionCells[`${r.x}-${r.y - 1}`] = true;
          visionCells[`${r.x}-${r.y + 1}`] = true;
          visionCells[`${r.x + 1}-${r.y - 1}`] = true;
          visionCells[`${r.x + 1}-${r.y}`] = true;
          visionCells[`${r.x + 1}-${r.y + 1}`] = true;
          break;
        case 4:
          visionCells[`${r.x + 1}-${r.y}`] = true;
          visionCells[`${r.x - 1}-${r.y}`] = true;
          visionCells[`${r.x}-${r.y + 1}`] = true;
          visionCells[`${r.x}-${r.y - 1}`] = true;
          break;
        default:
      }
    });

    let result = [];
    for (let y = 0; y < size; y++) {
      for (let x = 0; x < size; x++) {
        const s = {};
        if (visionCells[`${x}-${y}`]) {
          s.backgroundColor = '#ccc';
        }
        result.push(<div key={`${x}-${y}`} className="cell" style={s} />);
      }
    }
    return result;
  }
}
