/* global.js — shared utilities */

/**
 * Build a WebSocket URL from the current page location.
 * Works correctly behind nginx (http→ws, https→wss).
 */
function buildWsUrl(path) {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  return `${proto}://${location.host}${path}`;
}
