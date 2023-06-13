import React from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import 'regenerator-runtime/runtime'; // to use async/await (https://github.com/babel/babel/issues/9849)
import App from './App';
import store from './redux/store';

const root = createRoot(document.getElementById('app'));
root.render(
  <Provider store={store}>
    <App />
  </Provider>
);
