import type { HealthResponse, PortMapping, Settings, StatusResponse } from './types';

const invalidBrowserTokenMessage = 'invalid browser token';

function browserToken(): string {
  if (typeof document === 'undefined') {
    return '';
  }
  return document.querySelector<HTMLMetaElement>('meta[name="porto-browser-token"]')?.content ?? '';
}

function requestHeaders(init?: RequestInit): Headers {
  const headers = new Headers(init?.headers);
  headers.set('Content-Type', 'application/json');
  const token = browserToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  return headers;
}

function updateBrowserToken(token: string): void {
  if (typeof document === 'undefined' || !token) {
    return;
  }
  let meta = document.querySelector<HTMLMetaElement>('meta[name="porto-browser-token"]');
  if (!meta) {
    meta = document.createElement('meta');
    meta.name = 'porto-browser-token';
    document.head.appendChild(meta);
  }
  meta.content = token;
}

async function refreshBrowserToken(): Promise<boolean> {
  if (typeof window === 'undefined') {
    return false;
  }
  const res = await fetch(`/?porto_token_refresh=${Date.now()}`, {
    cache: 'no-store',
    headers: { Accept: 'text/html' },
  });
  if (!res.ok) {
    return false;
  }
  const html = await res.text();
  const doc = new DOMParser().parseFromString(html, 'text/html');
  const token = doc.querySelector<HTMLMetaElement>('meta[name="porto-browser-token"]')?.content ?? '';
  if (!token || token === browserToken()) {
    return false;
  }
  updateBrowserToken(token);
  return true;
}

async function request<T>(path: string, init?: RequestInit, allowTokenRefresh = true): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: requestHeaders(init),
  });
  if (!res.ok) {
    let errMsg = `${path} failed: ${res.status}`;
    try {
      const rawText = await res.text();
      if (rawText) {
        try {
          const errObj = JSON.parse(rawText);
          if (errObj && typeof errObj === 'object' && 'message' in errObj) {
            errMsg = String(errObj.message);
          } else {
            errMsg = rawText;
          }
        } catch {
          errMsg = rawText;
        }
      }
    } catch {
      errMsg = `${path} failed: ${res.status}`;
    }
    if (allowTokenRefresh && res.status === 401 && errMsg.includes(invalidBrowserTokenMessage) && await refreshBrowserToken()) {
      return request<T>(path, init, false);
    }
    throw new Error(errMsg);
  }
  return (await res.json()) as T;
}

export const api = {
  health: () => request<HealthResponse>('/api/health'),
  status: () => request<StatusResponse>('/api/status'),
  discover: () => request<StatusResponse>('/api/discover', { method: 'POST' }),
  openPort: (mapping: PortMapping) => request<StatusResponse>('/api/ports/open', {
    method: 'POST',
    body: JSON.stringify(mapping),
  }),
  closePort: (mapping: PortMapping) => request<StatusResponse>('/api/ports/close', {
    method: 'POST',
    body: JSON.stringify(mapping),
  }),
  getSettings: () => request<Settings>('/api/settings'),
  saveSettings: (settings: Settings) => request<{ ok: boolean }>('/api/settings', {
    method: 'POST',
    body: JSON.stringify(settings),
  }),
};
