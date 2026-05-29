import App from './App.svelte';
import './app.css';
import '@fontsource/material-symbols-outlined';

const app = new App({
  target: document.getElementById('app') as HTMLElement,
});

export default app;
