import React, { Component } from 'react';
import './App.css';
import Grid from './Grid';
import Robot from './Robot';
import Leaderboard from './Leaderboard';
import { connect } from 'react-refetch';

class App extends Component {
  render() {
    const { fetchState } = this.props;

    if (!fetchState.fulfilled) {
      return <div>Loading...</div>;
    }

    const state = fetchState.value;

    return (
      <div className="App">
        <header className="App-header">
          <div className="container">
            <h1>Round {state.round}</h1>
            <Leaderboard leaders={state.robots} />
          </div>
        </header>
        <div className="grid">
          <Grid size={state.grid} robots={state.robots} />
          {state.robots.map(r => (
            <Robot key={r.name} {...r} />
          ))}
        </div>
      </div>
    );
  }
}

export default connect(props => ({
  fetchState: {
    url: `/state`,
    refreshInterval: 1000
  }
  // fetchState: {
  //   value: {
  //     round: 1,
  //     grid: 16,
  //     robots: [
  //       { x: 1, y: 0, color: '#e6194b', name: 'JP', direction: 1, vision: 4, score: 1000 },
  //       { x: 5, y: 2, color: '#3cb44b', name: 'HP', direction: 3, vision: 8, score: 10 },
  //       { x: 10, y: 10, color: '#ff00ff', name: 'JW', direction: 3, vision: 4, score: 50 }
  //     ]
  //   }
  // }
}))(App);
