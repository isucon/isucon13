// import emojiData from '@emoji-mart/data';
import emojiData from '@emoji-mart/data/sets/14/twitter.json';
import { init as emojiInit } from 'emoji-mart';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App } from '~/App';
import '~/index.css';

emojiInit({ data: emojiData });

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
);
