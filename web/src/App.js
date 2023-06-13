import React from 'react';
import LogDisplay from './components/LogDisplay';

function App() {
  console.log('rendering in app.');
  return (
    <div className="App">
      <LogDisplay />
    </div>
  );
}

export default App;
