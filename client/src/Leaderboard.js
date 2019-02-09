import React, { Component } from 'react';

export default class Leaderboard extends Component {
  render() {
    const { leaders } = this.props;

    const result = [];
    leaders.forEach(function(a) {
      if (!this[a.name]) {
        this[a.name] = { name: a.name, score: 0 };
        result.push(this[a.name]);
      }
      this[a.name].score += a.score;
    }, Object.create(null));

    result.sort((a, b) => b.score - a.score);

    return (
      <table>
        <tbody>
          {result.slice(0, 8).map(r => (
            <tr key={r.name}>
              <td>{r.name}</td>
              <td>{r.score}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
}
