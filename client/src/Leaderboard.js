import React, { Component } from 'react';

export default class Leaderboard extends Component {
  render() {
    const { leaders } = this.props;

    leaders.sort((a, b) => b.score - a.score);

    return (
      <table>
        <tbody>
          {leaders.map(r => (
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
