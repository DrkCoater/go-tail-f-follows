import 'regenerator-runtime/runtime'; // to use async/await (https://github.com/babel/babel/issues/9849)
import React, { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import store from './redux/store';
import App from './App';

const root = createRoot(document.getElementById('app'));
root.render(
    <StrictMode>
        <Provider store={store}>
            <App />
        </Provider>
    </StrictMode>
);
