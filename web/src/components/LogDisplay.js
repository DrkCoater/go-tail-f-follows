import React, { useEffect, useRef, useState } from 'react';
export default function LogDisplayComponent() {
  const [logs, setLogs] = useState('');
  let ws = useRef(null);
  const clearLogs = () => {
    setLogs('');
  };
  useEffect(() => {
    ws.current = new WebSocket('ws://localhost:8000/ws');
    ws.current.onmessage = (event) => {
      setLogs(
        (prevLogs) => `${prevLogs}\n${event.data}`);
    };
    ws.current.onclose = () => console.log('ws closed');
    ws.current.onerror = () => console.log('ws error');
    return () => {
      ws.current.close();
    };
  }, []);
  return (
    <div>
      <textarea value={logs} readOnly style={{ width: '80vw', height: '80vh', display: 'block' }} />
      <button onClick={clearLogs}>Clear</button>
    </div>
  );
}
