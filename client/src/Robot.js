import React, { Component } from 'react';

export default class Robot extends Component {
  render() {
    const { x, y, color, name, direction } = this.props;
    const style = {
      left: 32 * x,
      top: 32 * y,
      borderBottomColor: color
    };
    const arrows = [['↑', 0, -1], ['→', 1, 0], ['↓', 0, 1], ['←', -1, 0]];
    return (
      <React.Fragment>
        <div className="robot" style={style}>
          {name}
        </div>
        <div
          className="arrow"
          style={{
            left: 32 * (x + arrows[direction][1]),
            top: 32 * (y + arrows[direction][2])
          }}
        >
          {arrows[direction][0]}
        </div>
      </React.Fragment>
    );
  }
}
